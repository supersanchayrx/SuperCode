package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"supercode/internal/config"
	"sync"
	"testing"
	"time"
)

type call struct {
	key   string
	model string
}

type fakeServer struct {
	mu     sync.Mutex
	calls  []call
	handle func(key, model string) (int, string)
}

func (f *fakeServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req ChatRequest
	_ = json.Unmarshal(body, &req)
	key := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	f.mu.Lock()
	f.calls = append(f.calls, call{key, req.Model})
	f.mu.Unlock()

	status, resp := f.handle(key, req.Model)
	w.WriteHeader(status)
	fmt.Fprint(w, resp)
}

func okReply(text string) string {
	return fmt.Sprintf(`{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":%q}}]}`, text)
}

func setup(t *testing.T, keys []string, handle func(key, model string) (int, string)) *fakeServer {
	t.Helper()

	f := &fakeServer{handle: handle}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)

	oldEndpoint, oldPool, oldNotify := endpoint, pool, Notify
	oldOrch, oldPlan, oldEdit := config.OrchestratorModel, config.PlanModel, config.EditModel
	t.Cleanup(func() {
		endpoint, pool, Notify = oldEndpoint, oldPool, oldNotify
		config.OrchestratorModel, config.PlanModel, config.EditModel = oldOrch, oldPlan, oldEdit
	})

	endpoint = srv.URL
	pool = newKeyPool(keys)
	Notify = func(string) {}
	config.OrchestratorModel, config.PlanModel, config.EditModel = "orch", "plan", "edit"

	return f
}

func ask(task Task) (Message, error) {
	return Complete(task, []Message{{Role: "user", Content: "hi"}}, nil)
}

func TestModelsForTask(t *testing.T) {
	setup(t, []string{"k1"}, nil)

	cases := map[Task]string{
		TaskChat:   "orch,plan,edit," + config.FallbackModel,
		TaskRouter: "orch,plan,edit," + config.FallbackModel,
		TaskPlan:   "plan,orch,edit," + config.FallbackModel,
		TaskEdit:   "edit,orch,plan," + config.FallbackModel,
	}

	for task, want := range cases {
		if got := strings.Join(ModelsFor(task), ","); got != want {
			t.Errorf("ModelsFor(%s) = %s, want %s", task, got, want)
		}
	}
}

func TestKeysRotateEvenly(t *testing.T) {
	f := setup(t, []string{"k1", "k2", "k3", "k4"}, func(key, model string) (int, string) {
		return 200, okReply("ok")
	})

	for i := 0; i < 8; i++ {
		if _, err := ask(TaskChat); err != nil {
			t.Fatal(err)
		}
	}

	counts := map[string]int{}
	for _, c := range f.calls {
		counts[c.key]++
		if c.model != "orch" {
			t.Errorf("chat used model %s, want orch", c.model)
		}
	}

	for _, k := range []string{"k1", "k2", "k3", "k4"} {
		if counts[k] != 2 {
			t.Errorf("%s used %d times, want 2 (counts %v)", k, counts[k], counts)
		}
	}
}

func TestRateLimitedKeySwitchesToNextKey(t *testing.T) {
	f := setup(t, []string{"k1", "k2"}, func(key, model string) (int, string) {
		if key == "k1" {
			return 429, `{"error":{"message":"Rate limit exceeded: free-models-per-min.","code":429,"metadata":{"provider_name":null}}}`
		}
		return 200, okReply("from k2")
	})

	msg, err := ask(TaskEdit)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "from k2" {
		t.Errorf("reply = %q", msg.Content)
	}
	if len(f.calls) != 2 || f.calls[1].model != "edit" {
		t.Errorf("calls = %v, want k1 then k2 on the edit model", f.calls)
	}

	if _, err := ask(TaskEdit); err != nil {
		t.Fatal(err)
	}
	if last := f.calls[len(f.calls)-1]; last.key != "k2" {
		t.Errorf("rate-limited k1 was used again: %v", f.calls)
	}
}

func TestRejectedKeyIsDisabled(t *testing.T) {
	setup(t, []string{"bad", "good"}, func(key, model string) (int, string) {
		if key == "bad" {
			return 401, `{"error":{"message":"User not found.","code":401}}`
		}
		return 200, okReply("ok")
	})

	if _, err := ask(TaskChat); err != nil {
		t.Fatal(err)
	}
	if pool.keys[0].disabled == "" {
		t.Error("key that got 401 was not disabled")
	}
}

func TestBrokenModelFallsThroughToBackup(t *testing.T) {
	f := setup(t, []string{"k1", "k2"}, func(key, model string) (int, string) {
		switch model {
		case "plan":
			return 404, `{"error":{"message":"No endpoints found for plan.","code":404}}`
		case "orch":
			return 502, `{"error":{"message":"Provider returned error","code":502,"metadata":{"provider_name":"X"}}}`
		case "edit":
			return 200, `{"error":{"message":"upstream crashed","code":500}}`
		}
		return 200, okReply("from backup")
	})

	msg, err := ask(TaskPlan)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Content != "from backup" {
		t.Errorf("reply = %q", msg.Content)
	}

	var models []string
	for _, c := range f.calls {
		models = append(models, c.model)
	}
	if got := strings.Join(models, ","); got != "plan,orch,edit,"+config.FallbackModel {
		t.Errorf("models tried = %s", got)
	}

	f.calls = nil
	if _, err := ask(TaskPlan); err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 || f.calls[0].model != config.FallbackModel {
		t.Errorf("broken models were not skipped on the next request: %v", f.calls)
	}
}

func TestUpstreamRateLimitIsAModelProblem(t *testing.T) {
	setup(t, []string{"k1"}, func(key, model string) (int, string) {
		if model == "orch" {
			return 429, `{"error":{"message":"orch is temporarily rate-limited upstream","code":429,"metadata":{"provider_name":"P"}}}`
		}
		return 200, okReply("ok")
	})

	if _, err := ask(TaskChat); err != nil {
		t.Fatal(err)
	}
	if !pool.keys[0].coolUntil.IsZero() {
		t.Error("an upstream (provider) rate limit cooled the API key")
	}
}

func TestEverythingFailsReturnsError(t *testing.T) {
	setup(t, []string{"k1"}, func(key, model string) (int, string) {
		return 400, `{"error":{"message":"bad request","code":400}}`
	})

	_, err := ask(TaskChat)
	if err == nil || !strings.Contains(err.Error(), "all models and API keys failed") {
		t.Errorf("err = %v", err)
	}
}

func TestKeyPerMinuteBudget(t *testing.T) {
	setup(t, []string{"k1"}, nil)

	k := pool.keys[0]
	now := time.Now()
	for i := 0; i < keyRequestsPerMinute; i++ {
		k.recent = append(k.recent, now.Add(-30*time.Second))
	}

	if pool.pickKey(now) != nil {
		t.Error("a key at its per-minute budget was picked")
	}
	if wait := pool.soonestReady(now); wait < 25*time.Second || wait > 35*time.Second {
		t.Errorf("soonestReady = %s, want about 30s", wait)
	}
	if pool.pickKey(now.Add(31*time.Second)) == nil {
		t.Error("key was not usable again after its oldest request aged out")
	}
}

func TestRouterDoesNotWait(t *testing.T) {
	setup(t, []string{"k1"}, func(key, model string) (int, string) {
		return 429, `{"error":{"message":"Rate limit exceeded","code":429}}`
	})

	start := time.Now()
	if _, err := ask(TaskRouter); err == nil {
		t.Fatal("expected an error")
	}
	if time.Since(start) > 5*time.Second {
		t.Error("router waited for a rate limit")
	}
}

func TestRetryAfterFromMetadata(t *testing.T) {
	reset := time.Now().Add(42 * time.Second).UnixMilli()
	body := fmt.Sprintf(`{"error":{"message":"Rate limit exceeded","code":429,"metadata":{"headers":{"X-RateLimit-Reset":"%d"}}}}`, reset)
	err := newAPIError(429, http.Header{}, []byte(body))

	if err.RetryAfter < 40*time.Second || err.RetryAfter > 43*time.Second {
		t.Errorf("RetryAfter = %s, want about 42s", err.RetryAfter)
	}
	if err.Upstream {
		t.Error("account rate limit marked as upstream")
	}
}
