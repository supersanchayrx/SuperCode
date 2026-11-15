package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"supercode/internal/agent"

	"github.com/spf13/cobra"
)

var modeFlag string

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Used to enter chat mode",
	RunE:  chat,
}

func init() {
	chatCmd.Flags().StringVarP(&modeFlag, "mode", "M", "auto", "auto,edit,plan,chat")
	rootCmd.AddCommand(chatCmd)
}

func chat(cmd *cobra.Command, args []string) error {
	mode, err := agent.ParseMode(modeFlag)
	if err != nil {
		return err
	}
	a := agent.New(mode)

	switch a.Mode {
	case agent.ModeChat:
		chatMode(a)

	default:
		chatMode(a)
	}

	return nil

}

func chatMode(agent *agent.Agent) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter a query. Switch modes using the /mode command")

	for {
		fmt.Printf("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}

		err := agent.Chat(input)

		if err != nil {
			fmt.Println("Error : ", err)
			continue
		}
	}

	fmt.Println("Exiting Chat mode ... BYEEE!!!")
}
