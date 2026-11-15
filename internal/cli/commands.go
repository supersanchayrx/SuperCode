package cli

import (
	"fmt"
	"strings"
	"supercode/internal/agent"
)

const helpText = `Commands:
  /mode <name>   switch mode (auto, chat, plan, edit)
  /mode          show the current mode
  /auto /chat /plan /edit   shortcuts for /mode <name>
  /help          show this help
  /exit, /quit   end the session`

func handleCommand(a *agent.Agent, input string) bool {

	fields := strings.Fields(input)

	cmd := strings.ToLower(fields[0])

	args := fields[1:]

	switch cmd {
	case "/exit", "/quit":
		return true
	case "/help":
		fmt.Println(helpText)
	case "/mode":
		{
			if len(args) == 0 {
				fmt.Printf("Current mode: %s\n", a.Mode)
				fmt.Println("usage: /mode <auto|chat|plan|edit>")
			} else {
				err := a.SetMode(args[0])
				if err != nil {
					fmt.Println("Can't switch mode:", err)
					return false
				}
				fmt.Println("Switched to", a.Mode, "mode")
			}

		}

	case "/chat", "/edit", "/plan", "/auto":
		modeString := strings.TrimPrefix(cmd, "/")
		if err := a.SetMode(modeString); err != nil {
			fmt.Printf("Failed to change modes: %s\n", err)
		} else {
			fmt.Println("Switched to", a.Mode, "mode")
		}

	default:
		fmt.Println("Unknown command:", cmd, "(for help type in '/help')")
	}

	return false
}
