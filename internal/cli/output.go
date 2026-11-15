package cli

import (
	"fmt"
	"time"
)

var outputQueue []string

func AddToOutputQueue() {

}

func streamText(text string, delay time.Duration) {
	for _, char := range text {
		fmt.Printf("%c", char)
		time.Sleep(delay)
	}
}
