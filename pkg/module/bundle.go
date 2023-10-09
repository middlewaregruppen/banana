package module

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/getsops/sops/v3/aes"
	"github.com/getsops/sops/v3/keys"

	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/age"
	syaml "github.com/getsops/sops/v3/stores/yaml"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func FromResMap(rm resmap.ResMap) *Bundle {
	return &Bundle{
		ResMap: rm,
	}
}

type Bundle struct {
	ResMap resmap.ResMap
	mod    Module
	opts   []BundleOpts

	// SOPS
	//sopsOutputStore, sopsInputStore *syaml.Store
	//sopsCipher aes.Cipher
}

type BundleOpts func(*Bundle) *Bundle

type BuildOpts struct {
	AgeRecipients []string
}

func (b *Bundle) FindByGVK(gvk GroupVersionKind) []*resource.Resource {
	res := b.ResMap.GetMatchingResourcesByAnyId(func(id resid.ResId) bool {
		if id.Group == gvk.Group && id.Kind == gvk.Kind && id.Version == gvk.Version {
			return true
		}
		return false
	})
	return res
}

func (b *Bundle) Flatten(w io.Writer) error {
	// Apply middleware
	for _, opt := range b.opts {
		opt(b)
	}
	d, err := b.ResMap.AsYaml()
	if err != nil {
		return err
	}
	_, err = w.Write(d)
	return err
}

func (b *Bundle) FlattenSecure(recipients []string, w io.Writer) error {
	buf := &bytes.Buffer{}
	err := b.Flatten(buf)
	if err != nil {
		return err
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

	//ncryptedRegex := strings.Join(keysToEncrypt, "|")

	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			KeyGroups: []sops.KeyGroup{groups},
			Version:   "v1.0.0",
			//UnencryptedRegex: "^(apiVersion|metadata|kind|type)$",
			//EncryptedRegex:   fmt.Sprintf("^(%s)", encryptedRegex),
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

func WithSOPS(recipients []string) BundleOpts {
	return func(b *Bundle) *Bundle {
		secretResources := b.FindByGVK(GroupVersionKind{"", "v1", "Secret"})
		for _, k := range b.mod.Secrets() {
			for _, secRes := range secretResources {
				_, err := secRes.Pipe(
					kyaml.Lookup("data", k.Key),
					kyaml.Set(kyaml.NewScalarRNode(k.Value)),
				)
				if err != nil {
					return nil
				}
				idSet := resource.MakeIdSet([]*resource.Resource{secRes})
				err = b.ResMap.ApplySmPatch(idSet, secRes)
				if err != nil {
					return nil
				}
			}
		}
		return b
	}
}

// ApplyURLs applies all the URLs defined in this module to the provided ResMap.
func WithURLs(s string) BundleOpts {
	return func(b *Bundle) *Bundle {
		// Create a list of ingress resources to transform
		ingressResources := b.FindByGVK(GroupVersionKind{"networking.k8s.io", "v1", "Ingress"})

		// Loop throught the ingresses found, change the host field, and finally apply the patch
		for i, ing := range ingressResources {
			_, err := ing.Pipe(
				kyaml.LookupCreate(kyaml.MappingNode, "spec", "rules", fmt.Sprint(i)),
				kyaml.SetField("host", kyaml.NewScalarRNode(s)),
			)
			if err != nil {
				return nil
			}
			idSet := resource.MakeIdSet(ingressResources)
			err = b.ResMap.ApplySmPatch(idSet, ing)
			if err != nil {
				return nil
			}
		}
		return b
	}
}

// ApplySecrets applies all secrets defined in this module to the provided ResMap.
// Searches through the given resmap for Secret resources, updating/adding secrets on this module.
// The Secret resource to update is determined by the secret key name itself.
// This function only adds or updates values in the Secret resource if the key matches that of the module.
func WithSecrets(secrets []Secret) BundleOpts {

	return func(b *Bundle) *Bundle {
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
					return nil
				}
				idSet := resource.MakeIdSet([]*resource.Resource{secRes})
				err = b.ResMap.ApplySmPatch(idSet, secRes)
				if err != nil {
					return nil
				}
			}
		}
		return b
	}
}
