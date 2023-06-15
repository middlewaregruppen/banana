package module

import (
	"io"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

var DefaultKustomizerOptions = &krusty.Options{
	Reorder:           krusty.ReorderOptionNone,
	AddManagedbyLabel: false,
	LoadRestrictions:  ktypes.LoadRestrictionsRootOnly,
	PluginConfig:      ktypes.DisabledPluginConfig(),
}

type kustomizeModule struct {
	resmap  resmap.ResMap
	name    string
	version string
	// opts   types.ModuleOpts
}

func (m kustomizeModule) Name() string {
	return m.name
}

func (m kustomizeModule) Version() string {
	return m.version
}

func (m kustomizeModule) Build(w io.Writer) error {

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
