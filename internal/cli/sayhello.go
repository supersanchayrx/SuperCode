package cli

import (
	"fmt"
	"strings"
	"supercode/internal/llm"

	"github.com/spf13/cobra"
)

var name string

var sayHelloCmd = &cobra.Command{
	Use:   "sayhello",
	Short: "Says Hello using the ai model",
	Run:   sayHello,
}

func init() {
	sayHelloCmd.Flags().StringVarP(&name, "name", "n", "", "name used to refer to user")
	rootCmd.AddCommand(sayHelloCmd)
}

func sayHello(cmd *cobra.Command, args []string) {
	sayHelloFromModel()
}

func sayHelloFromModel() {

	var promptText strings.Builder
	promptText.WriteString("Reply with a system booting like message and greet the user")
	if name != "" {
		promptText.WriteString(" and refer the user with the name - " + name)
	}

	finalPrompt := promptText.String()
	request, err := llm.BuildChatRequest("", finalPrompt)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	answer, err := llm.HandleRequest(request)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(answer.Content)
}
