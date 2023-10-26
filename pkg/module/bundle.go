package module

import (
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/getsops/sops/v3/aes"
	"github.com/getsops/sops/v3/keys"

	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/age"
	syaml "github.com/getsops/sops/v3/stores/yaml"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type Bundle struct {
	resources     []Resource
	resmap        resmap.ResMap
	mod           Module
	opts          []BundleOpts
	exportRootDir string
	recipients    []string
}

type BundleOpts func(*Bundle) error
type ExportOpts func(*Bundle) error

func (b *Bundle) Resources() []Resource {
	return b.resources
}

func (b *Bundle) FindByGVK(gvk GroupVersionKind) []*resource.Resource {
	res := b.resmap.GetMatchingResourcesByAnyId(func(id resid.ResId) bool {
		if id.Group == gvk.Group && id.Kind == gvk.Kind && id.Version == gvk.Version {
			return true
		}
		return false
	})
	return res
}

func (b *Bundle) Export(fs filesys.FileSystem, opts ...ExportOpts) error {

	// Apply opts
	for _, opt := range opts {
		opt(b)
	}

	root := b.mod.Name()
	if len(b.exportRootDir) > 0 {
		root = path.Join(b.exportRootDir, root)
	}

	err := fs.MkdirAll(root)
	if err != nil {
		return err
	}

	// Write each resource to file on fs
	for _, res := range b.Resources() {
		fname := strings.ToLower(path.Join(root, fmt.Sprintf("%s_%s.yaml", res.GetKind(), res.GetName())))
		dfile, err := fs.Create(fname)
		if err != nil {
			return err
		}
		defer dfile.Close()

		var out []byte

		out, err = res.Flatten()
		if err != nil {
			return err
		}

		if len(b.recipients) > 0 && res.GetKind() == "Secret" && res.GetApiVersion() == "v1" {
			out, err = res.FlattenSecure(
				b.recipients,
				b.mod.Secrets(),
			)
			if err != nil {
				return err
			}
		}

		_, err = dfile.Write(out)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Bundle) Flatten(w io.Writer) error {
	for _, res := range b.resmap.Resources() {
		d, err := res.AsYAML()
		if err != nil {
			return err
		}
		_, err = w.Write(d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bundle) FlattenSecure(recipients []string, w io.Writer) error {
	buf := &bytes.Buffer{}
	err := b.Flatten(buf)
	if err != nil {
		return err
	}

	// Skip encryption if no secrets exist in module
	if len(b.mod.Secrets()) <= 0 {
		if _, err := w.Write(buf.Bytes()); err != nil {
			return err
		}
		return nil
	}

	var keysToEncrypt []string
	for _, sec := range b.mod.Secrets() {
		keysToEncrypt = append(keysToEncrypt, sec.Key)
	}
	encrypted, err := encrypt(buf.Bytes(), recipients, keysToEncrypt)
	if err != nil {
		return err
	}
	_, err = w.Write(encrypted)
	return err
}

func encrypt(data []byte, recipients, keysToEncrypt []string) ([]byte, error) {
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

	encryptedRegex := strings.Join(keysToEncrypt, "|")

	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			KeyGroups:      []sops.KeyGroup{groups},
			Version:        "v1.0.0",
			EncryptedRegex: fmt.Sprintf("^(%s)", encryptedRegex),
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

func WithExportRootDir(root string) ExportOpts {
	return func(b *Bundle) error {
		b.exportRootDir = root
		return nil
	}
}

func WithResMap(rm resmap.ResMap) BundleOpts {
	return func(b *Bundle) error {
		b.resmap = rm
		b.resources = make([]Resource, len(b.resmap.Resources()))
		for i, res := range b.resmap.Resources() {
			b.resources[i].Resource = res
		}
		return nil
	}
}

func WithAgeRecipients(recipients []string) BundleOpts {
	return func(b *Bundle) error {
		b.recipients = recipients
		return nil
	}
}

// ApplyURLs applies all the URLs defined in this module to the provided ResMap.
func WithURLs(s string) BundleOpts {
	return func(b *Bundle) error {
		// Create a list of ingress resources to transform
		ingressResources := b.FindByGVK(GroupVersionKind{"networking.k8s.io", "v1", "Ingress"})

		// Loop throught the ingresses found, change the host field, and finally apply the patch
		for i, ing := range ingressResources {
			_, err := ing.Pipe(
				kyaml.LookupCreate(kyaml.MappingNode, "spec", "rules", fmt.Sprint(i)),
				kyaml.SetField("host", kyaml.NewScalarRNode(s)),
			)
			if err != nil {
				return err
			}
			idSet := resource.MakeIdSet(ingressResources)
			err = b.resmap.ApplySmPatch(idSet, ing)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// ApplySecrets applies all secrets defined in this module to the provided ResMap.
// Searches through the given resmap for Secret resources, updating/adding secrets on this module.
// The Secret resource to update is determined by the secret key name itself.
// This function only adds or updates values in the Secret resource if the key matches that of the module.
func WithSecrets(secrets []Secret) BundleOpts {
	return func(b *Bundle) error {
		// Create a list of Secret resources to transform
		secretResources := b.FindByGVK(GroupVersionKind{"", "v1", "Secret"})

		// Range over each secret. If the key matches that of a Secret resource then replace it's value with a strategic merge patch
		for _, k := range secrets {
			for _, secRes := range secretResources {
				_, err := secRes.Pipe(
					kyaml.Lookup("data", k.Key),
					kyaml.Set(kyaml.NewScalarRNode(k.Value)),
				)
				if err != nil {
					return err
				}
				idSet := resource.MakeIdSet([]*resource.Resource{secRes})
				err = b.resmap.ApplySmPatch(idSet, secRes)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// NewBundle returns a new Bundle for the given module and options provided
func NewBundle(m Module, opts ...BundleOpts) (*Bundle, error) {
	b := &Bundle{
		mod:  m,
		opts: opts,
	}
	// Apply opts
	for _, opt := range opts {
		err := opt(b)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}
