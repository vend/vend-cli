package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	cmd "github.com/vend/vend-cli/commands"
	"github.com/vend/vend-cli/pkg/messenger"
)

const uhOhLogo = `


  ____  __ __   _  ______     __  __
 / __ \/ // /  / |/ / __ \   / / / /
/ /_/ / _  /  /    / /_/ /  /_/ /_/ 
\____/_//_/  /_/|_/\____/  (_) (_)  
								   `

func SupressStackTrace() {
	if r := recover(); r != nil {
		fmt.Println(color.RedString(uhOhLogo))
		if exit, ok := r.(messenger.Exit); ok {
			fmt.Println(color.RedString("vendcli exited because an error occured"))
			fmt.Println("ERROR:  ", color.YellowString(fmt.Sprintf("%s", exit.Message)))
			os.Exit(exit.Code)
		}
		fmt.Println(color.RedString("vendcli exited unexpectedly"))
		fmt.Println(color.YellowString("please share the following stack trace in swarm"))
		panic(r) // not an Exit, bubble up
	}
}

func main() {
	defer SupressStackTrace()
	cmd.Execute()
}
