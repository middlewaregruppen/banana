package cmd

import (
	"os"

	"github.com/amimof/kmaint/cmd/create"
	"github.com/amimof/kmaint/cmd/version"

	"github.com/spf13/cobra"
)

func NewDefaultCommand() *cobra.Command {
	stdOut := os.Stdout
	c := &cobra.Command{
		Use:   "kmaint",
		Short: "The multi-purpose command line tool for maintaining application state",
		Long: `The multi-purpose command line tool for maintaining application state.
	kmaint is a simple CLI tool to:
	- Describe application configuration that emerges into Kustomize projects
	download and manage the Kubernetes Fury Distribution (KFD) modules
	`,
	}

	c.AddCommand(version.NewCmdVersion(stdOut))
	c.AddCommand(create.NewCmdCreate())
	return c
}
