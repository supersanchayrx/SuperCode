package cli

import (
	"fmt"

	"supercode/internal/config"

	"github.com/spf13/cobra"
)

var workDir string

var runCmd = &cobra.Command{
	Use:   "run [task]",
	Short: "Make this thing do sm coding",
	Long:  "Try giving this thing some coding work and it will try it's best to explore,edit,verify the code files (no guarantee of sucess)",
	Args:  cobra.ExactArgs(1), // will add more args later but keeping it simple for now by just requiring one arg i.e the task itself
	Run: func(cmd *cobra.Command, args []string) {
		task := args[0]
		cfg := config.Load()

		if cfg.APIKey == "" {
			fmt.Println("Put in the Api Key Bruh")
			return
		}

		fmt.Println("SuperCode 🦹")
		fmt.Printf("Working Dir: %s\n", workDir)
		fmt.Printf("Task: %s\n", task)
		fmt.Println()
		fmt.Println("agent loop shi") //implement this further
	},
}

func init() {
	runCmd.Flags().StringVarP(&workDir, "dir", "d", ".", "working directory for this task")
	rootCmd.AddCommand(runCmd)
}
