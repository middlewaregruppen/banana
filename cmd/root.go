package cmd

import (
	"os"

	"github.com/amimof/kmaint/cmd/create"
	"github.com/amimof/kmaint/cmd/version"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func NewDefaultCommand() *cobra.Command {
	fs := filesys.MakeFsOnDisk()
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
	c.AddCommand(create.NewCmdCreate(fs))
	return c
}
