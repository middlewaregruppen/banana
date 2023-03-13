package main

import (
	"os"

	"github.com/amimof/kmaint/cmd"
)

func main() {
	if err := cmd.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
