package module

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/aes"
	"github.com/getsops/sops/v3/age"
	"github.com/getsops/sops/v3/keys"
	syaml "github.com/getsops/sops/v3/stores/yaml"
	"github.com/middlewaregruppen/banana/api/types"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/krusty"
	ktypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var DefaultKustomizerOptions = &krusty.Options{
	Reorder:           krusty.ReorderOptionNone,
	AddManagedbyLabel: false,
	LoadRestrictions:  ktypes.LoadRestrictionsNone,
	PluginConfig: &ktypes.PluginConfig{
		BpLoadingOptions: ktypes.BploUseStaticallyLinked,
		HelmConfig: ktypes.HelmConfig{
			Enabled: true,
			Command: "helm",
		},
		PluginRestrictions: ktypes.PluginRestrictionsBuiltinsOnly,
		FnpLoadingOptions:  ktypes.DisabledPluginConfig().FnpLoadingOptions,
	},
}

type KustomizeModule struct {
	mod       types.Module
	fs        filesys.FileSystem
	prefix    string
	tmpFolder filesys.ConfirmedDir
	//resmap resmap.ResMap
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

func (m *KustomizeModule) Ref() string {
	return m.mod.Ref
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
	k := krusty.MakeKustomizer(DefaultKustomizerOptions)
	_, err := k.Run(m.fs, m.Name())
	return err
}

func (m *KustomizeModule) Secrets() []Secret {
	var secrets []Secret
	for _, s := range m.mod.Secrets {
		secrets = append(secrets, getSecretFromString(s))
	}
	return secrets
}

// getHostName returns a string that can be used as value in an ingress host field.
// This function parses the modules Host struct and builds a hostname value based on the params provided.
// Prefix & Wildcard fields on the Host struct will be prepended and appended to the string provided to this function.
func (m *KustomizeModule) Host() string {
	// Don't bother if hosts isn't defined
	if m.mod.Hosts == nil {
		return ""
	}

	// Delimiter is '-' by default
	delim := "-"
	if len(m.mod.Hosts.Delimiter) > 0 {
		delim = m.mod.Hosts.Delimiter
	}

	// HostName will always have highest priority because it's explicit
	if len(m.mod.Hosts.HostName) > 0 {
		return m.mod.Hosts.HostName
	}

	// Default to the ingress resource name
	names := strings.Split(m.Name(), "/")
	name := names[len(names)-1]

	if len(m.mod.Hosts.Prefix) > 0 {
		name = fmt.Sprintf("%s%s%s", m.mod.Hosts.Prefix, delim, name)
	}

	if len(m.mod.Hosts.Wildcard) > 0 {
		delim = "."
		name = fmt.Sprintf("%s%s%s", name, delim, m.mod.Hosts.Wildcard)
	}

	return name
}

// Takes a secret in the form of key=value and returns they key and value as two different return values
func getSecretFromString(s string) Secret {
	ss := strings.Split(s, "=")
	key, val := ss[0], ss[1]
	// If len is greater than 2 it means the secret value contains a '='
	// so wee need to account for that by concatinating everything after the first occurance.
	if len(ss) > 2 {
		val = s[len(key)+1:]
	}
	return Secret{Key: key, Value: val}
}

// ApplySecrets applies all secrets defined in this module to the provided ResMap.
// Searches through the given resmap for Secret resources, updating/adding secrets on this module.
// The Secret resource to update is determined by the secret key name itself.
// This function only adds or updates values in the Secret resource if the key matches that of the module.
// func (m *KustomizeModule) ApplySecrets(rm resmap.ResMap) error {

// 	// Create a list of Secret resources to transform
// 	secretResources := rm.GetMatchingResourcesByAnyId(func(id resid.ResId) bool {
// 		return id.Kind == "Secret"
// 	})

// 	// Range over each secret. If the key matches that of a Secret resource then replace it's value with a strategic merge patch
// 	secrets := m.Secrets()
// 	for _, k := range secrets {
// 		for _, secRes := range secretResources {
// 			_, err := secRes.Pipe(
// 				kyaml.Lookup("data", k.Key),
// 				kyaml.Set(kyaml.NewScalarRNode(k.Value)),
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			idSet := resource.MakeIdSet([]*resource.Resource{secRes})
// 			err = rm.ApplySmPatch(idSet, secRes)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	return nil
// }

// ApplyURLs applies all the URLs defined in this module to the provided ResMap.
// func (m *KustomizeModule) ApplyURLs(rm resmap.ResMap) error {

// 	// Create a list of ingress resources to transform
// 	ingressResources := rm.GetMatchingResourcesByAnyId(func(id resid.ResId) bool {
// 		if id.Group == "networking.k8s.io" && id.Kind == "Ingress" {
// 			return true
// 		}
// 		return false
// 	})

// 	// Loop throught the ingresses found, change the host field, and finally apply the patch
// 	for i, ing := range ingressResources {

// 		// Build the hostname we are going to use to patch the host field of the ingress resource
// 		hostName := m.getHostName(ing.GetName())

// 		_, err := ing.Pipe(
// 			kyaml.LookupCreate(kyaml.MappingNode, "spec", "rules", fmt.Sprint(i)),
// 			kyaml.SetField("host", kyaml.NewScalarRNode(hostName)),
// 		)
// 		if err != nil {
// 			return nil
// 		}
// 		idSet := resource.MakeIdSet(ingressResources)
// 		err = rm.ApplySmPatch(idSet, ing)
// 		if err != nil {
// 			return nil
// 		}
// 	}
// 	return nil
// }

// func (m *KustomizeModule) Bundle(opts BuildOpts) (Bundle, error) {
// 	var factory = provider.NewDefaultDepProvider().GetResourceFactory()

// 	// Encrypt secrets
// 	if len(opts.AgeRecipients) > 0 {
// 		yml, err := secRes.AsYAML()
// 		if err != nil {
// 			return err
// 		}
// 		eyml, err := m.BuildEncrypted(yml, opts.AgeRecipients)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	secRes.AsYAML()
// }

func (m *KustomizeModule) Build(w io.Writer) error {
	return nil
}

func (m *KustomizeModule) Bundle(opts ...BundleOpts) (*Bundle, error) {

	// Create kustomization file in tmp fs
	kustFile := fmt.Sprintf("/%s/kustomization.yaml", m.tmpFolder.String())
	kf, err := m.fs.Create(kustFile)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	_, err = kf.Write(b)
	if err != nil {
		return nil, err
	}

	k := krusty.MakeKustomizer(DefaultKustomizerOptions)
	rm, err := k.Run(m.fs, m.tmpFolder.String())
	if err != nil {
		return nil, err
	}

	bopts := make([]BundleOpts, len(opts)+1)
	bopts[0] = WithResMap(rm)
	for i, o := range opts {
		bopts[i+1] = o
	}
	return NewBundle(m,
		bopts...,
	)
}

func (m *KustomizeModule) BuildEncrypted(data []byte, recipients []string) ([]byte, error) {
	outputStore := &syaml.Store{}
	inputStore := &syaml.Store{}
	cipher := aes.NewCipher()

	branches, err := inputStore.LoadPlainFile(data)
	if err != nil {
		return nil, err
	}

	var ageMasterKeys []keys.MasterKey
	ageKeys, err := age.MasterKeysFromRecipients(strings.Join(recipients, ","))
	if err != nil {
		return nil, err
	}
	for _, k := range ageKeys {
		ageMasterKeys = append(ageMasterKeys, k)
	}
	var groups sops.KeyGroup
	groups = append(groups, ageMasterKeys...)

	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			KeyGroups: []sops.KeyGroup{groups},
			Version:   "v1.0.0",
		},
	}

	dataKey, errs := tree.GenerateDataKey()
	if len(errs) > 0 {
		err = fmt.Errorf("could not generate data key: %s", errs)
		return nil, err
	}

	unencryptedMac, err := tree.Encrypt(dataKey, cipher)
	if err != nil {
		return nil, err
	}
	tree.Metadata.LastModified = time.Now().UTC()

	tree.Metadata.MessageAuthenticationCode, err = cipher.Encrypt(unencryptedMac, dataKey, tree.Metadata.LastModified.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	return outputStore.EmitEncryptedFile(tree)
}

// NewKustomizeModule creates a new Kustomize implemented Module. It expects a filesystem, module type and a prefix.
// The prefix argument is a string that will prefix this module's name. This is so that the name module can be mapped to a Git repository,
// without having to reference the module by it's entire URL.
func NewKustomizeModule(fs filesys.FileSystem, mod types.Module, prefix string, tmpFolder filesys.ConfirmedDir) *KustomizeModule {
	return &KustomizeModule{
		fs:        fs,
		mod:       mod,
		prefix:    prefix,
		tmpFolder: tmpFolder,
	}
}
