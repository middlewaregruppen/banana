package create

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var (
	fileName string
)

func NewCmdCreate(fs filesys.FileSystem) *cobra.Command {
	c := &cobra.Command{
		Use:     "create",
		Aliases: []string{"init"},
		Short:   "Initialize a new kmaint configuration in the current directory",
		Long:    "",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if fs.Exists(fileName) {
				return fmt.Errorf("kustomization file already exists")
			}
			f, err := fs.Create(fileName)
			if err != nil {
				return err
			}
			f.Close()
			return nil
		},
	}
	c.Flags().StringVarP(
		&fileName,
		"filename",
		"f",
		"config.yaml",
		"The files that contain the configurations to apply.")
	return c
}
