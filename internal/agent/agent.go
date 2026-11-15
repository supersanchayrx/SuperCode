package agent

import (
	"fmt"
	"supercode/internal/llm"
	"supercode/internal/tools"
)

type Agent struct {
	Mode      Mode
	Confirm   func(question string) bool
	chatTools *tools.Registry
	editTools *tools.Registry
	history   []llm.Message
}

func New(mode string) (*Agent, error) {
	startMode, err := ParseMode(mode)

	if err != nil {
		return nil, err
	}

	llm.Notify = func(message string) {
		fmt.Println(message)
	}

	return &Agent{
		Mode:      startMode,
		chatTools: tools.NewRegistry(tools.ReadFile(), tools.ListFiles()),
		editTools: tools.NewRegistry(tools.ReadFile(), tools.ListFiles(), tools.EditFile(), tools.WriteFile(), tools.RunCommand()),
	}, nil
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

func (a *Agent) Status() []string {
	return llm.Status()
}
