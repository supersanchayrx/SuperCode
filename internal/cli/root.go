package cli

import (
	"fmt"
	"os"
	"supercode/internal/agent"

	"github.com/spf13/cobra"
)

var modeFlag string

var Version = "dev"

const modeCommandHelperText string = "starting mode: auto, chat, plan or edit"

var rootCmd = &cobra.Command{
	Use:               "supercode",
	Short:             "Not so smart coding agent",
	RunE:              startSession,
	PersistentPreRunE: loadConfig,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVarP(&modeFlag, "mode", "M", "auto", modeCommandHelperText) //last arg is helper text
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startSession(cmd *cobra.Command, args []string) error {
	a, err := agent.New(modeFlag)

	if err != nil {
		return err
	}

	runRepl(a)
	return nil

}
