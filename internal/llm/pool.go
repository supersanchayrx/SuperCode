package llm

import (
	"errors"
	"fmt"
	"supercode/internal/config"
	"time"
)

type Task string

const (
	TaskRouter Task = "router"
	TaskChat   Task = "chat"
	TaskPlan   Task = "plan"
	TaskEdit   Task = "edit"
)

const (
	keyRequestsPerMinute = 20
	defaultRateLimitWait = 60 * time.Second
	maxWait              = 90 * time.Second
	modelCooldown        = 2 * time.Minute
	missingModelCooldown = 30 * time.Minute
	maxRounds            = 3
)

var Notify = func(message string) {}

type apiKey struct {
	label     string
	value     string
	disabled  string
	coolUntil time.Time
	recent    []time.Time
}

type keyPool struct {
	keys           []*apiKey
	modelCoolUntil map[string]time.Time
}

type outcome int

const (
	succeeded outcome = iota
	modelFailed
	outOfKeys
	fatal
)

var pool *keyPool

func newKeyPool(values []string) *keyPool {
	p := &keyPool{modelCoolUntil: make(map[string]time.Time)}

	for i, v := range values {
		p.keys = append(p.keys, &apiKey{label: fmt.Sprintf("key %d", i+1), value: v})
	}

	return p
}

func getPool() *keyPool {
	if pool == nil {
		pool = newKeyPool(config.APIKeys)
	}
	return pool
}

func ModelsFor(task Task) []string {
	primary := config.OrchestratorModel

	switch task {
	case TaskPlan:
		primary = config.PlanModel
	case TaskEdit:
		primary = config.EditModel
	}

	var models []string
	seen := make(map[string]bool)

	for _, m := range []string{primary, config.OrchestratorModel, config.PlanModel, config.EditModel, config.FallbackModel} {
		if m != "" && !seen[m] {
			seen[m] = true
			models = append(models, m)
		}
	}

	return models
}

func Complete(task Task, messages []Message, tools []ToolDef) (Message, error) {
	p := getPool()

	if len(p.keys) == 0 {
		return Message{}, errors.New("no API keys configured")
	}

	models := ModelsFor(task)
	var lastErr error

	for round := 0; round < maxRounds; round++ {
		keysRanOut := false
		candidates := p.candidates(models, time.Now())

		for i, model := range candidates {
			msg, result, err := p.tryModel(model, messages, tools)

			switch result {
			case succeeded:
				if model != models[0] {
					Notify(fmt.Sprintf("(answered by %s)", model))
				}
				return msg, nil
			case fatal:
				return Message{}, err
			case outOfKeys:
				keysRanOut = true
			case modelFailed:
				if i+1 < len(candidates) {
					Notify(fmt.Sprintf("(%s failed: %v; trying %s)", model, err, candidates[i+1]))
				}
			}

			lastErr = err

			if keysRanOut {
				break
			}
		}

		if !keysRanOut || task == TaskRouter {
			break
		}

		wait := p.soonestReady(time.Now())
		if wait <= 0 || wait > maxWait {
			break
		}

		Notify(fmt.Sprintf("(every API key is rate-limited; waiting %s)", wait.Round(time.Second)))
		time.Sleep(wait)
	}

	if lastErr == nil {
		lastErr = errors.New("no model could be reached")
	}

	return Message{}, fmt.Errorf("all models and API keys failed, last error: %w", lastErr)
}

func (p *keyPool) tryModel(model string, messages []Message, tools []ToolDef) (Message, outcome, error) {
	var lastErr error

	for {
		key := p.pickKey(time.Now())

		if key == nil {
			if lastErr == nil {
				lastErr = errors.New("every API key is rate-limited or disabled")
			}
			return Message{}, outOfKeys, lastErr
		}

		key.recent = append(key.recent, time.Now())

		msg, err := send(key.value, model, messages, tools)

		if err == nil {
			return msg, succeeded, nil
		}

		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			return Message{}, fatal, err
		}

		lastErr = apiErr

		switch {
		case apiErr.Status == 401 || apiErr.Status == 402 || apiErr.Status == 403:
			key.disabled = apiErr.Message
			Notify(fmt.Sprintf("(%s was rejected: %s; not using it again this session)", key.label, apiErr.Message))

		case apiErr.Status == 429 && !apiErr.Upstream:
			wait := apiErr.RetryAfter
			if wait <= 0 {
				wait = defaultRateLimitWait
			}
			key.coolUntil = time.Now().Add(wait)
			Notify(fmt.Sprintf("(%s is rate-limited for %s; switching key)", key.label, wait.Round(time.Second)))

		default:
			switch {
			case apiErr.Status == 404:
				p.modelCoolUntil[model] = time.Now().Add(missingModelCooldown)
			case apiErr.Status == 429 || apiErr.Status >= 500:
				p.modelCoolUntil[model] = time.Now().Add(modelCooldown)
			}
			return Message{}, modelFailed, apiErr
		}
	}
}

func (p *keyPool) candidates(models []string, now time.Time) []string {
	var ready []string

	for _, m := range models {
		if now.After(p.modelCoolUntil[m]) {
			ready = append(ready, m)
		}
	}

	if len(ready) == 0 {
		return models[len(models)-1:]
	}

	return ready
}

func (p *keyPool) pickKey(now time.Time) *apiKey {
	var best *apiKey

	for _, k := range p.keys {
		k.prune(now)

		if !k.usable(now) {
			continue
		}

		if best == nil || len(k.recent) < len(best.recent) ||
			(len(k.recent) == len(best.recent) && k.lastUsed().Before(best.lastUsed())) {
			best = k
		}
	}

	return best
}

func (p *keyPool) soonestReady(now time.Time) time.Duration {
	soonest := time.Duration(-1)

	for _, k := range p.keys {
		if k.disabled != "" {
			continue
		}

		k.prune(now)
		wait := k.readyIn(now)

		if soonest < 0 || wait < soonest {
			soonest = wait
		}
	}

	return soonest
}

func (k *apiKey) prune(now time.Time) {
	cutoff := now.Add(-time.Minute)
	i := 0
	for i < len(k.recent) && !k.recent[i].After(cutoff) {
		i++
	}
	k.recent = k.recent[i:]
}

func (k *apiKey) usable(now time.Time) bool {
	return k.disabled == "" && !now.Before(k.coolUntil) && len(k.recent) < keyRequestsPerMinute
}

func (k *apiKey) readyIn(now time.Time) time.Duration {
	wait := k.coolUntil.Sub(now)

	if len(k.recent) >= keyRequestsPerMinute {
		if w := k.recent[0].Add(time.Minute).Sub(now); w > wait {
			wait = w
		}
	}

	if wait < 0 {
		return 0
	}

	return wait
}

func (k *apiKey) lastUsed() time.Time {
	if len(k.recent) == 0 {
		return time.Time{}
	}
	return k.recent[len(k.recent)-1]
}

func Status() []string {
	p := getPool()
	now := time.Now()

	lines := []string{
		fmt.Sprintf("models: chat/auto %s, plan %s, edit %s, backup %s",
			config.OrchestratorModel, config.PlanModel, config.EditModel, config.FallbackModel),
	}

	for _, k := range p.keys {
		k.prune(now)

		state := fmt.Sprintf("ready, %d requests in the last minute", len(k.recent))
		switch {
		case k.disabled != "":
			state = "disabled: " + k.disabled
		case !k.usable(now):
			state = fmt.Sprintf("rate-limited, ready in %s", k.readyIn(now).Round(time.Second))
		}

		lines = append(lines, fmt.Sprintf("%s (…%s): %s", k.label, lastChars(k.value, 4), state))
	}

	for model, until := range p.modelCoolUntil {
		if now.Before(until) {
			lines = append(lines, fmt.Sprintf("%s: skipped for %s after errors", model, until.Sub(now).Round(time.Second)))
		}
	}

	return lines
}

func lastChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
