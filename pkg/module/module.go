package module

import (
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/kustomize/api/krusty"
	ktypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"

	"github.com/middlewaregruppen/banana/api/types"
)

var (
	TemplateSuffix = ".tmpl"
)

type Module interface {
	Build(io.Writer) error
}

type kustomizeModule struct {
	opts types.ModuleOpts
}

func (m kustomizeModule) Build(w io.Writer) error {
	// Create the folder structure
	modulePath := fmt.Sprintf("%s/%s", "modules", m.opts.ModuleName())
	srcPath := fmt.Sprintf("%s/%s", "src", m.opts.ModuleName())

	// First we check if module exists
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return err
	}

	// Create folder structure
	err := os.MkdirAll(srcPath, os.ModePerm)
	if err != nil {
		return err
	}

	fsys := filesys.MakeFsOnDisk()

	// Build kustomization with krusty api
	k := krusty.MakeKustomizer(&krusty.Options{
		Reorder:           krusty.ReorderOptionNone,
		AddManagedbyLabel: false,
		LoadRestrictions:  ktypes.LoadRestrictionsRootOnly,
		PluginConfig:      ktypes.DisabledPluginConfig(),
	})

	// Run kustomization
	res, err := k.Run(fsys, modulePath)
	if err != nil {
		return err
	}

	// As Yaml output
	yml, err := res.AsYaml()
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

func LoadKustomizeModule(opts types.ModuleOpts) Module {
	return &kustomizeModule{opts: opts}
}

func Load(m types.ModuleOpts) Module {
	// Determine module implementation here. Right now we only support kustomize modules
	return LoadKustomizeModule(m)
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
