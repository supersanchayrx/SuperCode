package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"supercode/internal/agent"
)

func runRepl(a *agent.Agent) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Type a message, or /help for commands")

	//actual repl loop
	for {
		fmt.Printf("\n[%s]> ", a.Mode)
		if !scanner.Scan() {
			break
		} //token to show on left

		input := strings.TrimSpace(scanner.Text()) // take input

		//if input is empty skip
		if input == "" {
			continue
		}

		if strings.HasPrefix(input, "/") {
			if handleCommand(a, input) {
				break
			}
			continue
		}

		err := a.Handle(input) //every input gets

		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
	}

	fmt.Println("BYEEE!!!")
}
