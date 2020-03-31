package main

import (
	"fmt"
	"os"

	"github.com/KeisukeYamashita/slackduty/cmd"
)

func main() {
	execute()
}

func execute() {
	handleError(cmd.Execute())
}

func handleError(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, (fmt.Sprintf("%s\n", err.Error())))
	}
}
