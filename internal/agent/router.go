package agent

import (
	"strings"
	"supercode/internal/llm"
)

const routerPrompt = `Classify the user's message into exactly one category.
Reply with ONLY one word:
chat - a question, explanation or conversation
plan - asking to design, plan or break down work
edit - asking to write, change, fix or create code or files`

func classify(input string) Mode {
	err, answer := llm.HandleRequest(llm.BuildChatRequest(routerPrompt, input))

	if err != nil {
		return ModeChat
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "plan":
		return ModePlan
	case "edit":
		return ModeEdit
	}

	return ModeChat
}
