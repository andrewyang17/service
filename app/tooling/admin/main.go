package main

import (
	"fmt"
	"os"

	"github.com/andrewyang17/service/app/tooling/admin/commands"
)

var build = "develop"

func main() {
	err := commands.Migrate()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
