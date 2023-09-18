package module

import (
	"io"
	"net/url"
	"strings"

	"sigs.k8s.io/kustomize/api/loader"
)

var (
	TemplateSuffix = ".tmpl"
)

type Module interface {
	Version() string
	Name() string
	URL() string
	Namespace() string
	Components() []string
	Resolve() error
	Secrets() []string
	Build(io.Writer) error
	//Vendor(string, filesys.FileSystem) error
}

func moduleNameFromURL(urlstring string) (string, error) {
	s := urlstring
	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}
	//
	s = strings.Trim(u.Path, "/")
	if strings.Contains(u.Path, "//") {
		p := strings.Split(u.Path, "//")
		if len(p) >= 2 {
			s = p[1]
		}
	}
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

func gitRefFromSource(src string) (string, error) {
	s := ""
	u, err := url.Parse(src)
	if err != nil {
		return s, err
	}
	if ref := u.Query().Get("ref"); len(ref) > 0 {
		s = ref
	}
	return s, nil
}

func IsRemote(name string) bool {
	return loader.IsRemoteFile(name)
}
