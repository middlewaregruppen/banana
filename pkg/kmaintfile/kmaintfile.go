package kmaintfile

import (
	"github.com/amimof/kmaint/api/types"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type KmaintFile struct {
	fs filesys.FileSystem
}

// NewKmaintFile returns a new instance.
func NewKmaintFile(fs filesys.FileSystem) *KmaintFile {
	return &KmaintFile{fs: fs}
}

// // KmaintFile returns a new instance.
func (k *KmaintFile) Read(path string) (*types.KmaintFile, error) {
	data, err := k.fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var kf types.KmaintFile
	if err := yaml.Unmarshal(data, &kf); err != nil {
		return nil, err
	}

	return &kf, nil
}

func (k *KmaintFile) Write(kf *types.KmaintFile, path string) error {
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
