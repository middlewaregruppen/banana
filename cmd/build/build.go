package build

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

func NewCmdBuild(fs filesys.FileSystem, w io.Writer, prefix string) *cobra.Command {
	c := &cobra.Command{
		Use: "build",
		//Aliases: []string{""},
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

			//var b bytes.Buffer
			//writer := bufio.NewWriter(&b)

			// Init loader for loading modules
			l := module.NewLoader(fs)

			// Range over each module. A module is a structure of Go template files.
			// Following code will clone the folder structure of each module, generate
			// files in the structure using template definition.
			for _, m := range km.Modules {
				logrus.Debugf("building module %s holding %d component(s) \n", m.Name, len(m.Components))
				mod := l.Load(m, prefix)

				// Create module folder structure
				srcPath := fmt.Sprintf("%s/%s", "src", mod.Name())
				err = os.MkdirAll(srcPath, os.ModePerm)
				if err != nil {
					return err
				}
				// Assemble module and components
				if err = mod.Build(os.Stdout); err != nil {
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
	return c
}
