package build

import (
	"fmt"
	"io"
	"os"

	"github.com/middlewaregruppen/banana/pkg/bananafile"
	"github.com/middlewaregruppen/banana/pkg/git"
	"github.com/middlewaregruppen/banana/pkg/module"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	//"sigs.k8s.io/kustomize/api/krusty"

	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var (
	fileName string
	output   string
	age      []string
)

func NewCmdBuild(fs filesys.FileSystem, w io.Writer, prefix string) *cobra.Command {
	c := &cobra.Command{
		Use: "build",
		//Aliases: []string{""},
		Args:  cobra.ExactArgs(0),
		Short: "Builds sources from bnanana specification",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if !fs.Exists(fileName) {
				return fmt.Errorf("banana file not found")
			}
			kf := bananafile.NewBananaFile(fs)
			km, err := kf.Read(fileName)
			if err != nil {
				return err
			}

			// Init loader for loading modules
			tmpfs := filesys.MakeFsInMemory()
			l := module.NewLoader(tmpfs)

			// Range over each module, loading them through the loader and building them
			// to the provided writer
			for _, m := range km.Modules {
				logrus.Debugf("building module %s holding %d component(s) \n", m.Name, len(m.Components))
				mod := l.Load(m, prefix)
				logrus.Debugf("Will clone repo %s version %s using subdir %s into", mod.URL(), mod.Version(), mod.Name())

				// Setup the cloner and clone into temporary filesystem
				err := git.NewCloner(mod).Clone(tmpfs)
				if err != nil {
					return err
				}

				// Build the module
				if err = mod.Build(module.BuildOpts{AgeRecipients: age}, os.Stdout); err != nil {
					return err
				}
			}
			return err
		},
	}
	c.Flags().StringVarP(
		&fileName,
		"filename",
		"f",
		"banana.yaml",
		"The files that contain the configurations to apply.")
	c.Flags().StringVarP(
		&output,
		"output",
		"o",
		"stdout",
		"build banana specifiction to either stdout or filesystem",
	)
	c.Flags().StringArrayVar(
		&age,
		"age",
		[]string{},
		"comma separated list of age recipients [$SOPS_AGE_RECIPIENTS]",
	)
	return c
}
