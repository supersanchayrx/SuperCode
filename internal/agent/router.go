package agent

import (
	"strings"
	"supercode/internal/llm"
)

func classify(input string) Mode {
	answer, err := llm.Complete(llm.TaskRouter, []llm.Message{
		{Role: "system", Content: routerPrompt},
		{Role: "user", Content: input},
	}, nil)

	if err != nil {
		return ModeChat
	}

	switch strings.ToLower(strings.TrimSpace(answer.Content)) {
	case "plan":
		return ModePlan
	case "edit":
		return ModeEdit
	}

	return ModeChat
}
