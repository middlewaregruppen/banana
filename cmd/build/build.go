package build

import (
	"fmt"
	"path"

	"github.com/amimof/kmaint/pkg/kmaintfile"
	"github.com/spf13/cobra"

	//"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var (
	fileName             string
	defaultKustomization = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization`
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
			for _, m := range km.Modules {
				modulePath := fmt.Sprintf(fmt.Sprintf("%s/%s", "src", m.Name))
				err = fs.MkdirAll(modulePath)
				if err != nil {
					return err
				}
				err = fs.WriteFile(path.Join(modulePath, "kustomization.yaml"), []byte(defaultKustomization))
				if err != nil {
					return err
				}
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
