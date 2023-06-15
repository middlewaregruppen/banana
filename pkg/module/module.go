package module

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/middlewaregruppen/banana/api/types"
)

var (
	TemplateSuffix = ".tmpl"
)

type Module interface {
	Version() string
	Name() string
	Save(string) error
	Build(io.Writer) error
}

type Loader struct {
	fsys filesys.FileSystem
}

func NewLoader(fsys filesys.FileSystem) *Loader {
	return &Loader{
		fsys: fsys,
	}
}

// Parse parses a module by its name and returns a module instance
func (l *Loader) Parse(mod types.Module) (Module, error) {

	var err error
	var m Module

	// Try with Kustomize
	m, err = NewKustomizeModule(l.fsys, mod)
	if err == nil {
		logrus.Debugf("kustomize module detected: %s", mod.Name)
		return m, err
	}

	// Handle any errors that may have been encountered this far
	if err != nil {
		return nil, fmt.Errorf("unable to parse module due to an error: %s", err)
	}

	// Try with Helm
	return nil, fmt.Errorf("unable to recognise %s as a module using any of the supported module implementations", mod.Name)
}

func getModuleNameFromURL(urlstring string) (string, error) {
	u, err := url.Parse(urlstring)
	if err != nil {
		return "", err
	}
	res := strings.TrimLeft(u.Path, "/")
	return res, nil
}

func IsRemote(name string) bool {
	return loader.IsRemoteFile(name)
}

func WithParentOpts(km *types.BananaFile, m types.Module) types.ModuleOpts {
	newModule := m
	newModule.Opts["Name"] = km.Name
	newModule.Opts["Version"] = km.Version
	newModule.Opts["Metadata"] = km.MetaData
	newModule.Opts["Clusters"] = km.Clusters
	newModule.Opts["Module"] = m
	return newModule.Opts
}
