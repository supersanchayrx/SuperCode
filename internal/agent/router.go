package agent

import (
	"strings"
	"supercode/internal/llm"
)

func classify(input string) Mode {
	request, err := llm.BuildChatRequest(routerPrompt, input)

	if err != nil {
		return ModeChat
	}

	answer, err := llm.HandleRequest(request)

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
