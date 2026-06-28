package main

import (
	"os"

	"github.com/sonmezerekrem/atrisos/app/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
