package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"supercode/internal/config"
	"supercode/internal/llm"

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

		//converting working dir to absolute directory path
		absDir, err := filepath.Abs(workDir)
		if err != nil {
			fmt.Printf("Invalid directory path \nERR: %s\n", err)
			return
		}

		info, err := os.Stat(absDir)
		if err != nil {
			fmt.Printf("Path does not exists \nERR: %s\n", err)
			return
		}

		if !info.IsDir() {
			fmt.Printf("This is not a directory \nDir: %s\n", absDir)
			return
		}

		workDir = absDir

		fmt.Printf("The working directory was resolved to be : %s\n", workDir)

		if cfg.APIKey == "" {
			fmt.Println("ERR: Put in the Api Key Bruh")
			return
		}

		fmt.Println("SuperCode 🦹")
		fmt.Printf("Working Dir: %s\n", workDir)
		fmt.Printf("Task: %s\n", task)
		fmt.Println()

		//Creating the LLM Client
		client := llm.NewClient(cfg)
		messages := []llm.Message{
			{Role: "user", Content: task},
		}

		var err2 error
		var reply string

		fmt.Println("THINKING...")
		if cfg.Stream {
			reply, err2 = client.ChatStream(messages)
		} else {
			reply, err2 = client.Chat(messages)
		}

		if err != nil {
			fmt.Printf("ERR: %s\n", err2)
			return
		}

		//fmt.Printf("%s\n", reply)
		_ = reply
	},
}

func init() {
	runCmd.Flags().StringVarP(&workDir, "dir", "d", ".", "working directory for this task")
	rootCmd.AddCommand(runCmd)
}
