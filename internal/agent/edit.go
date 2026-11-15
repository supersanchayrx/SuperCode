package agent

import "supercode/internal/llm"

func (a *Agent) Edit(input string) error {
	_, err := a.runLoop(llm.TaskEdit, editPrompt, input, a.editTools)
	return err
}
