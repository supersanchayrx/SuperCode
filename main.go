package main

import (
	"fmt"
	"supercode/internal/cli"
	"supercode/internal/config"
)

func main() {
	fmt.Println("Hello Reading api key -- ")
	config.LoadKeys()

	cli.Execute()
}
