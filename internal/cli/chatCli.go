package cli

import (
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Used to enter chat mode",
	RunE:  startSession,
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
