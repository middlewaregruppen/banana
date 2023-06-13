package bananafile

import (
	"github.com/middlewaregruppen/banana/api/types"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type BananaFile struct {
	fs filesys.FileSystem
}

// NewBananaFile returns a new instance.
func NewBananaFile(fs filesys.FileSystem) *BananaFile {
	return &BananaFile{fs: fs}
}

// // BananaFile returns a new instance.
func (k *BananaFile) Read(path string) (*types.BananaFile, error) {
	data, err := k.fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var kf types.BananaFile
	if err := yaml.Unmarshal(data, &kf); err != nil {
		return nil, err
	}

	return &kf, nil
}

func (k *BananaFile) Write(kf *types.BananaFile, path string) error {
	d, err := yaml.Marshal(kf)
	if err != nil {
		return err
	}
	err = k.fs.WriteFile(path, d)
	if err != nil {
		return nil
	}
	return nil
}
