package build

import (
	"fmt"

	"github.com/amimof/kmaint/pkg/kmaintfile"
	"github.com/amimof/kmaint/pkg/module"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	//"sigs.k8s.io/kustomize/api/krusty"

	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var (
	fileName string
)

func NewCmdBuild(fs filesys.FileSystem) *cobra.Command {
	c := &cobra.Command{
		Use: "build",
		//Aliases: []string{""},
		Short: "Builds kustmizations from konf",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if !fs.Exists(fileName) {
				return fmt.Errorf("kmaint file not found")
			}
			kf := kmaintfile.NewKmaintFile(fs)
			km, err := kf.Read(fileName)
			if err != nil {
				return err
			}

			// Range over each module. A module is a structure of Go template files.
			// Following code will clone the folder structure of each module, generate
			// files in the structure using template definition.
			for _, m := range km.Modules {
				logrus.Infof("Building module %s\n", m.Name)
				mod := module.Load(module.WithParentOpts(km, m))
				if err = mod.Build(); err != nil {
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
		"kmaint.yaml",
		"The files that contain the configurations to apply.")
	return c
}
