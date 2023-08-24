package save

import (
	"fmt"
	"io"
	"os"

	"github.com/middlewaregruppen/banana/pkg/bananafile"
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

func NewCmdSave(fs filesys.FileSystem, w io.Writer, prefix string) *cobra.Command {
	c := &cobra.Command{
		Use: "save",
		//Aliases: []string{""},
		Short: "Fetch remove module and save locally",
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
			l := module.NewLoader(fs)

			// Range over each module. A module is a structure of Go template files.
			// Following code will clone the folder structure of each module, generate
			// files in the structure using template definition.
			for _, m := range km.Modules {
				logrus.Infof("preparing module %s\n", m.Name)
				mod := l.Load(m, prefix)

				// Create module folder structure
				srcPath := fmt.Sprintf("%s/%s", "src", mod.Name())
				err = os.MkdirAll(srcPath, os.ModePerm)
				if err != nil {
					return err
				}
				// Save to disk
				logrus.Infof("saving module %s (%s) to %s", mod.Name(), mod.Version(), srcPath)
				if err := mod.Save("src"); err != nil {
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
		"save banana specifiction to either stdout or filesystem",
	)
	return c
}
