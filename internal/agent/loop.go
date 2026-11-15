package agent

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"supercode/internal/llm"
	"supercode/internal/tools"
)

const (
	maxSteps       = 20
	maxLabelValue  = 60
	maxResultShown = 200
)

func (a *Agent) runLoop(task llm.Task, systemPrompt, input string, reg *tools.Registry) (string, error) {
	messages := make([]llm.Message, 0, len(a.history)+2)
	messages = append(messages, llm.Message{Role: "system", Content: systemPrompt})
	messages = append(messages, a.history...)
	messages = append(messages, llm.Message{Role: "user", Content: input})

	defs := toolDefs(reg)

	for step := 0; step < maxSteps; step++ {
		reply, err := llm.Complete(task, messages, defs)

		if err != nil {
			return "", err
		}

		if reply.Role == "" {
			reply.Role = "assistant"
		}

		messages = append(messages, reply)

		text := strings.TrimSpace(reply.Content)

		if len(reply.ToolCalls) == 0 {
			fmt.Println(text)
			a.saveHistory(messages[1:])
			return text, nil
		}

		if text != "" {
			fmt.Println(text)
		}

		for _, call := range reply.ToolCalls {
			messages = append(messages, llm.Message{
				Role:       "tool",
				ToolCallID: call.ID,
				Content:    a.runTool(reg, call),
			})
		}
	}

	return "", fmt.Errorf("stopped after %d steps without a final answer", maxSteps)
}

func (a *Agent) runTool(reg *tools.Registry, call llm.ToolCall) string {
	name := call.Function.Name
	args := call.Function.Arguments
	label := describeCall(name, args)

	fmt.Println("→", label)

	if tool, ok := reg.Get(name); ok && tool.NeedsApproval {
		if tool.Preview != nil {
			preview, err := tool.Preview(args)

			if err != nil {
				fmt.Println("  error:", err)
				return "error: " + err.Error()
			}

			fmt.Println(indent(preview))
		}

		if !a.confirm(fmt.Sprintf("Allow %s?", name)) {
			fmt.Println("  declined")
			return "The user declined this action. Do not retry it; ask the user what they want instead."
		}
	}

	result, err := reg.Run(name, args)

	if err != nil {
		fmt.Println("  error:", err)
		return "error: " + err.Error()
	}

	fmt.Println(" ", firstLine(result, maxResultShown))
	return result
}

func (a *Agent) confirm(question string) bool {
	if a.Confirm == nil {
		return false
	}
	return a.Confirm(question)
}

func toolDefs(reg *tools.Registry) []llm.ToolDef {
	var defs []llm.ToolDef

	for _, t := range reg.List() {
		defs = append(defs, llm.ToolDef{
			Type: "function",
			Function: llm.FunctionDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	return defs
}

func describeCall(name, args string) string {
	var fields map[string]any

	if err := json.Unmarshal([]byte(args), &fields); err != nil || len(fields) == 0 {
		return name + "()"
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+shorten(fmt.Sprint(fields[k]), maxLabelValue))
	}

	return name + "(" + strings.Join(parts, ", ") + ")"
}

func indent(s string) string {
	return "    " + strings.ReplaceAll(s, "\n", "\n    ")
}

func firstLine(s string, max int) string {
	s = strings.TrimSpace(s)
	lines := strings.Count(s, "\n") + 1

	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}

	s = shorten(s, max)

	if lines > 1 {
		s += fmt.Sprintf(" (+%d more lines)", lines-1)
	}

	return s
}

func shorten(s string, max int) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "⏎")

	r := []rune(s)
	if len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}
