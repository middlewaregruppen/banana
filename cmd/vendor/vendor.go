package vendor

import (
	"fmt"
	"io"

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
)

func NewCmdVendor(fs filesys.FileSystem, w io.Writer, prefix string) *cobra.Command {
	c := &cobra.Command{
		Use: "vendor",
		//Aliases: []string{""},
		Short: "Vendors sources from banana specification",
		Long:  "",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !fs.Exists(fileName) {
				return fmt.Errorf("banana file not found")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			kf := bananafile.NewBananaFile(fs)
			km, err := kf.Read(fileName)
			if err != nil {
				return err
			}

			// Make fs on disk for vendoring modules
			fs := filesys.MakeFsOnDisk()

			// Init loader for loading modules
			l := module.NewLoader(fs)

			// Range over each module. A module is a structure of Go template files.
			// Following code will clone the folder structure of each module, generate
			// files in the structure using template definition.
			for _, m := range km.Modules {
				logrus.Debugf("vendoring module %s holding %d component(s) \n", m.Name, len(m.Components))
				mod := l.Load(m, prefix)
				dstPath := "src"
				logrus.Debugf("Will clone repo %s version %s using subdir %s into %s", mod.URL(), mod.Version(), mod.Name(), dstPath)
				err := git.NewCloner(mod,
					git.WithCloneSubDir(mod.Name()),
					git.WithTargetPath(dstPath),
				).Clone(fs)
				if err != nil {
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
		"vendor banana specifiction to either stdout or filesystem",
	)
	return c
}
