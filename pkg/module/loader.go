package module

import (
	"log"

	"github.com/middlewaregruppen/banana/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type Loader struct {
	fsys filesys.FileSystem
	TmpFolder filesys.ConfirmedDir
	mods []Module
}

// Parse parses a module by its name and returns a module instance
func (l *Loader) Load(mod types.Module, prefix string) Module {

	// Try with Kustomize
	m := NewKustomizeModule(l.fsys, mod, prefix, l.TmpFolder)
	l.mods = append(l.mods, m)
	return m
	// err = m.Resolve()
	// if err == nil {
	// 	logrus.Debugf("kustomize module detected: %s", mod.Name)
	// 	l.mods = append(l.mods, m)
	// 	return m, err
	// }

	// // Handle any errors that may have been encountered this far
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to parse module due to an error: %s", err)
	// }

	// // Try with Helm
	// return nil, fmt.Errorf("unable to recognise %s as a module using any of the supported module implementations", mod.Name)
}

func NewLoader(fsys filesys.FileSystem) *Loader {
	tmpfs, err := filesys.NewTmpConfirmedDir()
	if err != nil {
		log.Fatal(err)
	}

	return &Loader{
		fsys: fsys,
		TmpFolder: tmpfs,
	}
}
