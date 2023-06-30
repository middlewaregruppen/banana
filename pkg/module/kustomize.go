package module

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/middlewaregruppen/banana/api/types"
	"github.com/middlewaregruppen/banana/pkg/git"
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
	url     string
	version string
	fs      filesys.FileSystem
}

func (m *KustomizeModule) Name() string {
	n, _ := moduleNameFromURL(m.name)
	return n
}

func (m *KustomizeModule) Version() string {
	return m.version
}

func (m *KustomizeModule) URL() string {
	return m.url
}

func (m *KustomizeModule) Build(w io.Writer) error {

	k := krusty.MakeKustomizer(DefaultKustomizerOptions)
	res, err := k.Run(m.fs, m.URL())
	if err != nil {
		return err
	}
	m.resmap = res

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

func (m *KustomizeModule) Save(rootpath string) error {
	cloneURL, err := gitURLFromSource(m.url)
	if err != nil {
		return err
	}
	clonePath, err := moduleNameFromURL(m.url)
	if err != nil {
		return err
	}
	logrus.Debugf("Will clone %s %s", cloneURL, clonePath)

	tmpfs := filesys.MakeFsInMemory()
	err = git.Clone(tmpfs, cloneURL, clonePath)
	if err != nil {
		return err
	}

	// Walk over the tmp fs in which we have a cloned repo, attempting to
	// copy files and folders from the tmp fs to local disk
	return tmpfs.Walk(clonePath, func(rel string, info os.FileInfo, err error) error {

		// Skip if we get a dir
		if info.IsDir() {
			return err
		}

		// Create dir structure
		dstName := path.Join(rootpath, rel)
		err = os.MkdirAll(filepath.Dir(dstName), os.ModePerm)
		if err != nil {
			return err
		}

		// Stat file and check if it's a regular file
		if !info.Mode().IsRegular() {
			return err
		}

		// Open file for reading
		srcFile, err := tmpfs.Open(rel)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Create destination file which we will be writing to
		dstFile, err := os.Create(dstName)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// Copy content from source to destination file
		b, err := io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
		logrus.Debugf("Wrote %d bytes to %s", b, dstName)
		return nil
	})
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
		moduleName, _ = moduleNameFromURL(mod.Name)
		moduleURL = fmt.Sprintf("%s?ref=%s", strings.TrimRight(mod.Name, "/"), mod.Version)
	}

	logrus.Debugf("module source is %s", moduleURL)

	return &KustomizeModule{
		version: mod.Version,
		name:    moduleName,
		fs:      fs,
		url:     moduleURL,
	}, nil
}
