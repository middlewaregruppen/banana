package build

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/amimof/kmaint/pkg/kmaintfile"
	"github.com/spf13/cobra"

	//"sigs.k8s.io/kustomize/api/krusty"
	"text/template"

	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var (
	templateSuffix = ".tmpl"
	fileName       string
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

				// Create the folder structure
				modulePath := fmt.Sprintf("%s/%s", "modules", m.Name)
				srcPath := fmt.Sprintf("%s/%s", "src", m.Name)
				err = os.MkdirAll(srcPath, os.ModePerm)
				if err != nil {
					return err
				}

				// Walk the folder structure and attempt to find template files
				err = filepath.Walk(modulePath, func(rel string, info os.FileInfo, err error) error {

					// Ignore directories
					if info.IsDir() {
						return err
					}

					// Replace leading path "modules/" with "src/" and create folder structure within it
					dirs := strings.Split(rel, string(os.PathSeparator))
					dirs[0] = "src"
					dstName := path.Join(dirs...)
					dstDir := filepath.Dir(dstName)
					err = os.MkdirAll(dstDir, os.ModePerm)
					if err != nil {
						return err
					}

					// Remove template suffix from file name
					if strings.HasSuffix(dstName, templateSuffix) {
						dstName = path.Join(dstDir, info.Name()[:len(info.Name())-len(templateSuffix)])
					}

					// Stat file and check if it's a regular file
					srcStat, err := os.Stat(rel)
					if err != nil {
						return err
					}
					if !srcStat.Mode().IsRegular() {
						return err
					}

					// Open file for reading
					srcFile, err := os.Open(rel)
					if err != nil {
						return err
					}
					defer srcFile.Close()

					// Create destination file which we will be writing to
					dstFile, err := os.Create(dstName)
					if err != nil {
						return err
					}
					defer dstFile.Close()

					// We only care about files ending in .tmpl when templating
					if !strings.HasSuffix(rel, templateSuffix) {
						_, err = io.Copy(dstFile, srcFile)
						if err != nil {
							return err
						}
					}

					// Parse template injecting variables into it
					tmpl, err := template.ParseFiles(rel)
					if err != nil {
						return err
					}

					// Write output to file
					err = tmpl.Execute(dstFile, m.Opts)
					if err != nil {
						return err
					}
					return nil
				})

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
