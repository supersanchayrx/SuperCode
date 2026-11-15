package agent

import (
	"strings"
	"supercode/internal/llm"
	"testing"
)

func turn(user string, replySize int) []llm.Message {
	return []llm.Message{
		{Role: "user", Content: user},
		{Role: "assistant", ToolCalls: []llm.ToolCall{{ID: "c1", Type: "function", Function: llm.FunctionCall{Name: "read_file", Arguments: `{"path":"x"}`}}}},
		{Role: "tool", ToolCallID: "c1", Content: strings.Repeat("x", replySize)},
		{Role: "assistant", Content: "done"},
	}
}

func TestSaveHistoryKeepsSmallHistory(t *testing.T) {
	a := &Agent{}
	var msgs []llm.Message
	msgs = append(msgs, turn("first", 10)...)
	msgs = append(msgs, turn("second", 10)...)

	a.saveHistory(msgs)

	if a.HistoryTurns() != 2 || len(a.history) != 8 {
		t.Errorf("turns = %d, messages = %d, want 2 and 8", a.HistoryTurns(), len(a.history))
	}
}

func TestSaveHistoryDropsOldestWholeTurns(t *testing.T) {
	a := &Agent{}
	var msgs []llm.Message
	msgs = append(msgs, turn("first", maxHistoryChars)...)
	msgs = append(msgs, turn("second", 10)...)

	a.saveHistory(msgs)

	if a.history[0].Role != "user" || a.history[0].Content != "second" {
		t.Fatalf("history starts with %+v, want the second user message", a.history[0])
	}
	if a.HistoryTurns() != 1 {
		t.Errorf("turns = %d, want 1", a.HistoryTurns())
	}
}

func TestSaveHistoryKeepsLastTurnEvenIfLarge(t *testing.T) {
	a := &Agent{}
	a.saveHistory(turn("only", maxHistoryChars*2))

	if a.HistoryTurns() != 1 {
		t.Errorf("turns = %d, want 1", a.HistoryTurns())
	}
}

func TestClearHistory(t *testing.T) {
	a := &Agent{}
	a.saveHistory(turn("first", 10))
	a.ClearHistory()

	if a.HistoryTurns() != 0 || len(a.history) != 0 {
		t.Error("history is not empty after ClearHistory")
	}
}

func TestDescribeCall(t *testing.T) {
	cases := map[string]string{
		`{"path":"main.go"}`:         "read_file(path=main.go)",
		`{"b":"2","a":"1"}`:          "read_file(a=1, b=2)",
		`{}`:                         "read_file()",
		`not json`:                   "read_file()",
		`{"content":"line1\nline2"}`: "read_file(content=line1⏎line2)",
	}

	for args, want := range cases {
		if got := describeCall("read_file", args); got != want {
			t.Errorf("describeCall(%s) = %q, want %q", args, got, want)
		}
	}
}

func TestToolDefsKeepsOrder(t *testing.T) {
	a, err := New("edit")
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, d := range toolDefs(a.editTools) {
		names = append(names, d.Function.Name)
		if d.Type != "function" || len(d.Function.Parameters) == 0 {
			t.Errorf("tool def %s is incomplete", d.Function.Name)
		}
	}

	if got := strings.Join(names, ","); got != "read_file,list_files,edit_file,write_file,run_command" {
		t.Errorf("edit tools = %s", got)
	}
}
