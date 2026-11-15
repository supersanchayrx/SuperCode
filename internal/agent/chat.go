package agent

import (
	"fmt"
	"supercode/internal/llm"
)

func (a *Agent) Chat(input string) error {

	request, err := llm.BuildChatRequest(chatPrompt, input)

	if err != nil {
		return err
	}

	answer, err := llm.HandleRequest(request)

	if err != nil {
		return err
	}

	fmt.Println(answer.Content)

	return nil
}
