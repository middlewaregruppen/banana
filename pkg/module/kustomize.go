package module

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/middlewaregruppen/banana/api/types"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	ktypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var DefaultKustomizerOptions = &krusty.Options{
	Reorder:           krusty.ReorderOptionNone,
	AddManagedbyLabel: false,
	LoadRestrictions:  ktypes.LoadRestrictionsRootOnly,
	PluginConfig:      ktypes.DisabledPluginConfig(),
}

type KustomizeModule struct {
	resmap  resmap.ResMap
	name    string
	version string
	fs      filesys.FileSystem
}

func (m *KustomizeModule) Name() string {
	return m.name
}

func (m *KustomizeModule) Version() string {
	return m.version
}

func (m *KustomizeModule) Build(w io.Writer) error {

	// As Yaml output
	yml, err := m.resmap.AsYaml()
	if err != nil {
		return err
	}

	// Write to writer
	_, err = w.Write(yml)
	if err != nil {
		return err
	}
	return err
}

func (m *KustomizeModule) Save(filepath string) error {
	moduleSrc := fmt.Sprintf("%s/%s", "modules", m.Name())
	files, err := m.fs.ReadDir(moduleSrc)
	if err != nil {
		return err
	}
	for _, f := range files {
		b, err := m.fs.ReadFile(path.Join(moduleSrc, f))
		if err != nil {
			return err
		}
		moduleDst := fmt.Sprintf("%s/%s", filepath, f)
		logrus.Debugf("Writing to %s", moduleDst)
		err = m.fs.WriteFile(moduleDst, b)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewKustomizeModule(fs filesys.FileSystem, mod types.Module) (*KustomizeModule, error) {

	// Prepend "modules/" if module is local
	moduleName := mod.Name
	moduleURL := mod.Name
	if !IsRemote(moduleURL) {
		moduleURL = fmt.Sprintf("%s/%s", "modules", mod.Name)
	}

	// Append ref in URL if version is provided
	if IsRemote(moduleURL) && len(mod.Version) > 0 {
		moduleName, _ = getModuleNameFromURL(mod.Name)
		moduleURL = fmt.Sprintf("%s?ref=%s", strings.TrimRight(mod.Name, "/"), mod.Version)
	}

	logrus.Debugf("module source is %s", moduleURL)

	k := krusty.MakeKustomizer(DefaultKustomizerOptions)
	res, err := k.Run(fs, moduleURL)
	if err != nil {
		return nil, err
	}

	return &KustomizeModule{
		version: mod.Version,
		name:    moduleName,
		resmap:  res,
		fs:      fs,
	}, nil
}
