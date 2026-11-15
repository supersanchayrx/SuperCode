package main

import (
	"fmt"
	"os"
	"supercode/internal/cli"
	"supercode/internal/config"
)

func main() {
	fmt.Println("Hello Reading api key -- ")

	if _, err := config.LoadKeys(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cli.Execute()
}
