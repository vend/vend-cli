package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	cmd "github.com/vend/vend-cli/commands"
	"github.com/vend/vend-cli/pkg/messenger"
)

func SupressStackTrace() {
	if r := recover(); r != nil {
		if exit, ok := r.(messenger.Exit); ok {
			fmt.Println(color.RedString("\n\nvendcli exited because an error occured:"))
			fmt.Println(color.YellowString(fmt.Sprintf("%s", exit.Message)))
			os.Exit(exit.Code)
		}
		panic(r) // not an Exit, bubble up
	}
}

func main() {
	defer SupressStackTrace()
	cmd.Execute()
}
