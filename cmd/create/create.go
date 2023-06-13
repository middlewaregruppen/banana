package create

import (
	"fmt"

	"github.com/middlewaregruppen/banana/api/types"
	"github.com/middlewaregruppen/banana/pkg/bananafile"
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
		Short:   "Initialize a new banana configuration in the current directory",
		Long:    "",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if fs.Exists(fileName) {
				return fmt.Errorf("banana file already exists")
			}
			f, err := fs.Create(fileName)
			if err != nil {
				return err
			}
			f.Close()
			kf := bananafile.NewBananaFile(fs)
			km, err := kf.Read(fileName)
			if err != nil {
				return err
			}
			km.Kind = "Banana"
			km.APIVersion = "banana.io/v1alpha1"
			km.Name = ""
			km.Modules = []types.Module{}
			err = kf.Write(km, fileName)
			if err != nil {
				return nil
			}
			return nil
		},
	}
	c.Flags().StringVarP(
		&fileName,
		"filename",
		"f",
		"banana.yaml",
		"The files that contain the configurations to apply.")
	return c
}
