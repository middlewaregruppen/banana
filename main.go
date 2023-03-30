package main

import (
	"os"

	"github.com/amimof/kmaint/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.NewDefaultCommand().Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	os.Exit(0)
}
