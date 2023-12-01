package build

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
	//age      []string
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

			// Setup filesystem for exported bundles
			outfs := filesys.MakeFsOnDisk()

			// Init loader for loading modules
			l := module.NewLoader(fs)

			// Range over each module, loading them through the loader and building them
			// to the provided writer
			for _, m := range km.Modules {
				logrus.Debugf("building module %s holding %d component(s) \n", m.Name, len(m.Components))
				mod := l.Load(m, prefix)
				logrus.Debugf("Will clone repo %s version %s using subdir %s into", mod.URL(), mod.Version(), mod.Name())

				// Setup the cloner and clone into temporary filesystem
				opts := git.WithTargetPath(l.TmpFolder.String())
				err := git.NewCloner(mod, opts).Clone(fs)
				if err != nil {
					return err
				}

				// Init opts
				opts := []module.BundleOpts{
					module.WithSecrets(mod.Secrets()),
					module.WithURLs(mod.Host()),
				}

				// Use sops encryption if age recipients is provided
				if km.Age != nil && len(km.Age.Recipients) > 0 {
					opts = append(opts, module.WithAgeRecipients(km.Age.Recipients))
				}

				// Bundle the module
				bun, err := mod.Bundle(opts...)
				if err != nil {
					return err
				}

				// Write to disk
				err = bun.Export(outfs, module.WithExportRootDir("src/"))
				if err != nil {
					return err
				}

				// Build encrypted & flattened module to stdout
				// if km.Age != nil && len(km.Age.Recipients) > 0 {
				// 	if err = bun.FlattenSecure(km.Age.Recipients, os.Stdout); err != nil {
				// 		return err
				// 	}
				// 	return nil
				// }

				// // Build flattened module to stdout
				// if err = bun.Flatten(os.Stdout); err != nil {
				// 	return err
				// }
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
