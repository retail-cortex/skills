package main

import (
	"os"

	"github.com/retail-cortex/skills/internal/commands"
)

func main() {
	exitCode := commands.Execute(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(exitCode)
}
