package agent

import (
	"errors"
	"fmt"
	"strings"
	"supercode/internal/llm"
	"supercode/internal/tools"
)

const planFile = "PLAN.md"

func (a *Agent) Plan(input string) error {
	plan, err := a.runLoop(llm.TaskPlan, planPrompt, input, a.chatTools)

	if err != nil {
		return err
	}

	plan = strings.TrimSpace(plan)

	if plan == "" {
		return errors.New("the model returned an empty plan")
	}

	if tools.FileExists(planFile) && !a.confirm(fmt.Sprintf("%s already exists. Overwrite it?", planFile)) {
		fmt.Printf("Kept the existing %s\n", planFile)
		return nil
	}

	if _, err := tools.SaveFile(planFile, plan+"\n"); err != nil {
		return err
	}

	fmt.Printf("Saved the plan to %s. Use /edit to start on it.\n", planFile)
	return nil
}
