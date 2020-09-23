package main

import (
	"fmt"

	"github.com/devops-simba/redns/definitions"
)

func main() {
	app := CreateApplicationFromArgs()
	err := app.Execute()
	if err != nil {
		fmt.Printf("%s Operation failed: %v", definitions.Console.Write(definitions.Red, "[ERR]"), err)
	}
	app.Execute()
}
