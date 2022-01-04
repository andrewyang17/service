package main

import (
	"fmt"
	"os"

	"github.com/andrewyang17/service/app/tooling/admin/commands"
)

func main() {
	err := commands.GenKey()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
