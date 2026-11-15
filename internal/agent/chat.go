package agent

import "supercode/internal/llm"

func (a *Agent) Chat(input string) error {
	_, err := a.runLoop(llm.TaskChat, chatPrompt, input, a.chatTools)
	return err
}
