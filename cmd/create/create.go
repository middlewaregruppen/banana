package create

import (
	"fmt"

	"github.com/amimof/kmaint/api/types"
	"github.com/amimof/kmaint/pkg/kmaintfile"
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
				return fmt.Errorf("kmaint file already exists")
			}
			f, err := fs.Create(fileName)
			if err != nil {
				return err
			}
			f.Close()
			kf := kmaintfile.NewKmaintFile(fs)
			km, err := kf.Read(fileName)
			if err != nil {
				return err
			}
			km.Kind = "Konf"
			km.APIVersion = "konf.io/v1alpha1"
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
		"kmaint.yaml",
		"The files that contain the configurations to apply.")
	return c
}
