package agent

import "fmt"

type Agent struct {
	Mode Mode
}

func New(mode Mode) *Agent {
	return &Agent{Mode: mode}
}

func (a *Agent) Handle(input string) error {
	mode := a.Mode //local copy of agent's mode so auto doesn't go perma off
	if mode == ModeAuto {
		mode = classify(input)
		fmt.Printf("(auto → %s)\n", mode)
	}

	switch mode {
	case ModePlan:
		return a.Plan(input)
	case ModeEdit:
		return a.Edit(input)
	default:
		return a.Chat(input)
	}
}
