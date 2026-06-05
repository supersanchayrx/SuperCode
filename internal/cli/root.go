package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

//cobra command is a struct
var rootCmd = &cobra.Command{
	Use:   "supercode",
	Short: "Not so smart coding agent :/",
	Long:  "SuperCode is go based implementation of cli based coding agents like opencode/nightcode \nPretty much the same thing with a few *less* features \nUse at your own risk! 🫡",
}

func Execute() {
	err := rootCmd.Execute() 
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
