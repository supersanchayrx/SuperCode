package agent

import "supercode/internal/llm"

const maxHistoryChars = 120000

func (a *Agent) ClearHistory() {
	a.history = nil
}

func (a *Agent) HistoryTurns() int {
	turns := 0
	for _, m := range a.history {
		if m.Role == "user" {
			turns++
		}
	}
	return turns
}

func (a *Agent) saveHistory(messages []llm.Message) {
	a.history = append([]llm.Message(nil), messages...)

	for historySize(a.history) > maxHistoryChars {
		next := -1
		for i := 1; i < len(a.history); i++ {
			if a.history[i].Role == "user" {
				next = i
				break
			}
		}

		if next == -1 {
			return
		}

		a.history = a.history[next:]
	}
}

func historySize(messages []llm.Message) int {
	size := 0
	for _, m := range messages {
		size += len(m.Content)
		for _, call := range m.ToolCalls {
			size += len(call.Function.Arguments)
		}
	}
	return size
}
