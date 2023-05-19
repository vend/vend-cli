package main

import (
	"github.com/vend/govend/vend"
	cmd "github.com/vend/vend-cli/commands"
)

func main() {
	defer vend.SupressStackTrace()
	cmd.Execute()
}
