package cmd

import (
	"os"

	"github.com/middlewaregruppen/banana/cmd/build"
	"github.com/middlewaregruppen/banana/cmd/create"
	"github.com/middlewaregruppen/banana/cmd/vendor"
	"github.com/middlewaregruppen/banana/cmd/version"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var v string
var builtinModulePrefix string

func NewDefaultCommand() *cobra.Command {
	fs := filesys.MakeFsOnDisk()
	//fs := filesys.MakeFsInMemory()
	stdOut := os.Stdout
	c := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		Use:           "banana",
		Short:         "The multi-purpose command line tool for maintaining application state",
		Long: `The multi-purpose command line tool for maintaining application state.
	banana is a simple CLI tool to:
	- Describe application configuration that emerges into Kustomize projects
	`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			//logrus.SetOutput(os.Stdout)
			lvl, err := logrus.ParseLevel(v)
			if err != nil {
				return err
			}
			logrus.SetLevel(lvl)
			return nil
		},
	}

	// Setup flags
	c.PersistentFlags().StringVarP(
		&v,
		"v",
		"v",
		"info",
		"number for the log level verbosity (debug, info, warn, error, fatal, panic)")
	c.PersistentFlags().StringVar(
		&builtinModulePrefix,
		"builtin-module-prefix",
		"https://github.com/middlewaregruppen/banana-modules",
		"Prefix used for builtin modules. For example ingress/nginx or monitoring/grafana",
	)

	// Setup sub-commands
	c.AddCommand(version.NewCmdVersion(stdOut))
	c.AddCommand(create.NewCmdCreate(fs))
	c.AddCommand(build.NewCmdBuild(fs, stdOut, builtinModulePrefix))
	c.AddCommand(vendor.NewCmdVendor(fs, stdOut, builtinModulePrefix))

	return c
}
