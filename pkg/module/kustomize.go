package module

import (
	"fmt"
	"io"
	"strings"

	"github.com/middlewaregruppen/banana/api/types"
	"github.com/middlewaregruppen/banana/pkg/git"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	ktypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

var DefaultKustomizerOptions = &krusty.Options{
	Reorder:           krusty.ReorderOptionNone,
	AddManagedbyLabel: false,
	LoadRestrictions:  ktypes.LoadRestrictionsNone,
	PluginConfig:      ktypes.DisabledPluginConfig(),
}

type KustomizeModule struct {
	mod    types.Module
	fs     filesys.FileSystem
	prefix string
}

// Name returns a human readable version of this module.
// Returns m.Version if there is a value assigned. Otherwise
// the URL will be used to construct a value
func (m *KustomizeModule) Name() string {
	n := m.mod.Name
	if IsRemote(m.mod.Name) {
		n, _ = moduleNameFromURL(m.mod.Name)
	}
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
	u := m.prefix
	if IsRemote(m.mod.Name) {
		u, _ = gitURLFromSource(m.mod.Name)
	}
	return u
}

func (m *KustomizeModule) Namespace() string {
	return m.mod.Namespace
}

func (m *KustomizeModule) Components() []string {
	return m.mod.Components
}

func (m *KustomizeModule) Resolve() error {
	tmpfs := filesys.MakeFsInMemory()
	cloner := git.NewCloner(
		m.URL(),
		git.WithCloneTag(m.Version()),
	)
	err := cloner.Clone(tmpfs)
	if err != nil {
		return err
	}

	k := krusty.MakeKustomizer(DefaultKustomizerOptions)
	_, err = k.Run(tmpfs, m.Name())
	return err
}

func (m *KustomizeModule) Secrets() []string {
	return m.mod.Secrets
}

// getHostName returns a string that can be used as value in an ingress host field.
// This function parses the modules Host struct and builds a hostname value based on the params provided.
// Prefix & Wildcard fields on the Host struct will be prepended and appended to the string provided to this function.
func (m *KustomizeModule) getHostName(n string) string {
	// Don't bother if hosts isn't defined
	if m.mod.Host == nil {
		return ""
	}

	// Delimiter is '-' by default
	delim := "-"
	if len(m.mod.Host.Delimiter) > 0 {
		delim = m.mod.Host.Delimiter
	}

	// HostName will always have highest priority because it's explicit
	if len(m.mod.Host.HostName) > 0 {
		return m.mod.Host.HostName
	}

	// Default to the ingress resource name
	name := n

	if len(m.mod.Host.Prefix) > 0 {
		name = fmt.Sprintf("%s%s%s", m.mod.Host.Prefix, delim, name)
	}

	if len(m.mod.Host.Wildcard) > 0 {
		name = fmt.Sprintf("%s%s%s", name, delim, m.mod.Host.Wildcard)
	}

	return name
}

// ApplySecrets applies all the URLs defined in this module to the provided ResMap.
func (m *KustomizeModule) ApplySecrets(rm resmap.ResMap) error {
	// Create a list of ingress resources to transform
	secretResources := rm.GetMatchingResourcesByAnyId(func(id resid.ResId) bool {
		if id.Kind == "Secret" {
			fmt.Printf("%+v\n", id)
			return true
		}
		return false
	})

	secrets := m.Secrets()
	for _, k := range secrets {
		keyvalue := strings.Split(k, "=")
		n, v := keyvalue[0], keyvalue[1]
		// Loop throught the ingresses found, change the host field, and finally apply the patch
		for _, sec := range secretResources {
			res, err := sec.Pipe(
				kyaml.Lookup("data", n),
				kyaml.Set(kyaml.NewScalarRNode(v)),
			)
			if res == nil {
				return nil
			}
			if err != nil {
				return err
			}
			fmt.Println(res.String())
			fmt.Println(sec.String())
			idSet := resource.MakeIdSet([]*resource.Resource{sec})
			err = rm.ApplySmPatch(idSet, sec)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
// ApplyURLs applies all the URLs defined in this module to the provided ResMap.
func (m *KustomizeModule) ApplyURLs(rm resmap.ResMap) error {

	// Create a list of ingress resources to transform
	ingressResources := rm.GetMatchingResourcesByAnyId(func(id resid.ResId) bool {
		if id.Group == "networking.k8s.io" && id.Kind == "Ingress" {
			return true
		}
		return false
	})

	// Loop throught the ingresses found, change the host field, and finally apply the patch
	for i, ing := range ingressResources {

		// Build the hostname we are going to use to patch the host field of the ingress resource
		hostName := m.getHostName(ing.GetName())

		_, err := ing.Pipe(
			kyaml.LookupCreate(kyaml.MappingNode, "spec", "rules", fmt.Sprint(i)),
			kyaml.SetField("host", kyaml.NewScalarRNode(hostName)),
		)
		if err != nil {
			return nil
		}
		idSet := resource.MakeIdSet(ingressResources)
		err = rm.ApplySmPatch(idSet, ing)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (m *KustomizeModule) Build(w io.Writer) error {

	// Create kustomization file in tmp fs
	kf, err := m.fs.Create("kustomization.yaml")
	if err != nil {
		return err
	}
	defer kf.Close()

	// Compose the kustomiation file and encode it into yaml
	content := ktypes.Kustomization{
		TypeMeta: ktypes.TypeMeta{
			Kind:       ktypes.KustomizationKind,
			APIVersion: ktypes.KustomizationVersion,
		},
		Namespace:  m.Namespace(),
		Resources:  []string{m.Name()},
		Components: []string{},
	}

	// Clone every component into tmp fs
	for _, c := range m.Components() {
		cName := fmt.Sprintf("%s/%s", m.Name(), c)
		content.Components = append(content.Components, cName)
	}

	b, err := yaml.Marshal(&content)
	if err != nil {
		return err
	}

	_, err = kf.Write(b)
	if err != nil {
		return err
	}

	k := krusty.MakeKustomizer(DefaultKustomizerOptions)
	res, err := k.Run(m.fs, ".")
	if err != nil {
		return err
	}

	//m.resmap = res

	//fmt.Printf("\n\n%+v\n\n", ingressResources)
	err = m.ApplyURLs(res)
	if err != nil {
		return err
	}
	err = m.ApplySecrets(res)
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

func NewKustomizeModule(fs filesys.FileSystem, mod types.Module, prefix string) *KustomizeModule {
	return &KustomizeModule{
		fs:     fs,
		mod:    mod,
		prefix: prefix,
	}
}
