package module

import (
	"io"
	"net/url"
	"strings"

	"sigs.k8s.io/kustomize/api/loader"

	"github.com/middlewaregruppen/banana/api/types"
)

var (
	TemplateSuffix = ".tmpl"
)

type Module interface {
	Version() string
	Name() string
	URL() string
	Save(string) error
	Build(io.Writer) error
}

func moduleNameFromURL(urlstring string) (string, error) {
	s := urlstring
	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}
	if strings.Contains(u.Path, "//") {
		p := strings.Split(u.Path, "//")
		if len(p) >= 2 {
			s = p[1]
		}
	}
	//res := strings.TrimLeft(u.Path, "/")
	return s, nil
}

func gitURLFromSource(src string) (string, error) {
	s := src
	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}
	if strings.Contains(u.Path, "//") {
		p := strings.Split(u.Path, "//")
		if len(p) >= 1 {
			s = strings.Replace(s, u.Path, p[0], 1)
		}
	}
	return s, nil
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
