package agent

import "fmt"

type Mode string

const (
	ModeAuto Mode = "auto"
	ModeEdit Mode = "edit"
	ModePlan Mode = "plan"
	ModeChat Mode = "chat"
)

func ParseMode(s string) (Mode, error) {
	switch m := Mode(s); m {
	case ModeAuto, ModeChat, ModeEdit, ModePlan:
		return m, nil
	}

	return "", fmt.Errorf("unknown mode %q (use auto, chat, plan or edit)", s)
}
