package module

import (
	"io"
	"net/url"
	"strings"

	"sigs.k8s.io/kustomize/api/loader"
)

type Module interface {
	Version() string
	Ref() string
	Name() string
	URL() string
	Namespace() string
	Components() []string
	Resolve() error
	Secrets() []Secret
	Host() string
	Build(io.Writer) error
	Bundle(...BundleOpts) (*Bundle, error)
}

type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
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
