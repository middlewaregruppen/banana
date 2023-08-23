package module

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/middlewaregruppen/banana/api/types"
	"github.com/middlewaregruppen/banana/pkg/git"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	ktypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var DefaultKustomizerOptions = &krusty.Options{
	Reorder:           krusty.ReorderOptionNone,
	AddManagedbyLabel: false,
	LoadRestrictions:  ktypes.LoadRestrictionsNone,
	PluginConfig:      ktypes.DisabledPluginConfig(),
}

type KustomizeModule struct {
	mod    types.Module
	resmap resmap.ResMap
	fs     filesys.FileSystem
}

// Name returns a human readable version of this module.
// Returns m.Version if there is a value assigned. Otherwise
// the URL will be used to construct a value
func (m *KustomizeModule) Name() string {
	n, _ := moduleNameFromURL(m.mod.Name)
	return n
}

func (m *KustomizeModule) Version() string {
	if len(m.mod.Version) > 0 {
		return m.mod.Version
	}
	v, _ := gitRefFromSource(m.mod.Name)
	return v
}

func (m *KustomizeModule) URL() string {
	u, _ := gitURLFromSource(m.mod.Name)
	return u
}

func (m *KustomizeModule) Components() []types.Component {
	return m.mod.Components
}

func (m *KustomizeModule) Resolve() error {
	tmpfs := filesys.MakeFsInMemory()
	err := git.Clone(tmpfs, m.URL(), m.Version(), ".")
	if err != nil {
		return err
	}
	k := krusty.MakeKustomizer(DefaultKustomizerOptions)
	_, err = k.Run(tmpfs, m.Name())
	return err
}

func (m *KustomizeModule) Build(w io.Writer) error {

	// Create a surface area for the kustomization
	tmpfs := filesys.MakeFsInMemory()

	// Create kustomization file in tmp fs
	kf, err := tmpfs.Create("kustomization.yaml")
	if err != nil {
		return err
	}
	kf.Close()

	fmt.Println("exists?", tmpfs.Exists("kustomization.yaml"))
	// Compose the kustomiation file and encode it into yaml
	content := ktypes.Kustomization{
		TypeMeta: ktypes.TypeMeta{
			Kind:       ktypes.KustomizationKind,
			APIVersion: ktypes.KustomizationVersion,
		},
		Resources:  []string{m.Name()},
		Components: []string{},
	}

	// Clone module into tmp fs
	err = git.Clone(tmpfs, m.URL(), m.Version(), m.Name())
	if err != nil {
		return err
	}

	// Clone every component into tmp fs
	for _, c := range m.Components() {
		cName, err := moduleNameFromURL(c.Name)
		if err != nil {
			return err
		}
		cURL, err := gitURLFromSource(c.Name)
		if err != nil {
			return err
		}
		logrus.Debugf("cloning %s (%s) to %s", c.Name, c.Version, cName)
		err = git.Clone(tmpfs, cURL, c.Version, cName)
		if err != nil {
			return err
		}
		content.Components = append(content.Components, cName)
	}

	b, err := yaml.Marshal(&content)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", string(b))

	// _, err = kf.Write(b)
	// if err != nil {
	// 	return err
	// }

	err = tmpfs.WriteFile("/kustomization.yaml", b)
	if err != nil {
		return err
	}

	k := krusty.MakeKustomizer(DefaultKustomizerOptions)
	res, err := k.Run(tmpfs, ".")
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
	cloneURL := m.URL()
	clonePath := m.Name()
	cloneTag := m.Version()
	logrus.Debugf("Will clone %s version %s into %s", cloneURL, cloneTag, clonePath)

	tmpfs := filesys.MakeFsInMemory()
	err := git.Clone(tmpfs, cloneURL, cloneTag, clonePath)
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

func NewKustomizeModule(fs filesys.FileSystem, mod types.Module) *KustomizeModule {
	return &KustomizeModule{
		fs:  fs,
		mod: mod,
	}
}
