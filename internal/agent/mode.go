package agent

import (
	"fmt"
	"strings"
)

type Mode string

const (
	ModeAuto Mode = "auto"
	ModeEdit Mode = "edit"
	ModePlan Mode = "plan"
	ModeChat Mode = "chat"
)

func ParseMode(s string) (Mode, error) {
	switch m := Mode(strings.ToLower(s)); m {
	case ModeAuto, ModeChat, ModeEdit, ModePlan:
		return m, nil
	}

	return "", fmt.Errorf("unknown mode %q (use auto, chat, plan or edit)", s)
}

func (a *Agent) SetMode(name string) error {
	mode, err := ParseMode(name)

	if err != nil {
		return err
	}

	a.Mode = mode
	return nil
}
