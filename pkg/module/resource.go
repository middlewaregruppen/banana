package module

import (
	"bytes"
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/resource"
)

type Resource struct {
	*resource.Resource
}

func (r *Resource) FileName() string {
	return strings.ToLower(fmt.Sprintf("%s_%s.yaml", r.GetKind(), r.GetName()))
}

// FlattenSecure flattenes the resource returning a byte array containing a YAML representation of the resource.
func (r *Resource) Flatten() ([]byte, error) {
	d, err := r.AsYAML()
	if err != nil {
		return nil, err
	}
	return d, nil
}

// FlattenSecure flattenes the resource returning a byte array containing a YAML representation of the resource.
// If the resource is a v1.Secret, then the output will be encrypted using sops.
func (r *Resource) FlattenSecure(recipients []string, secs []Secret) ([]byte, error) {
	buf := &bytes.Buffer{}
	b, err := r.Flatten()
	if err != nil {
		return nil, err
	}

	var keysToEncrypt []string
	for _, sec := range secs {
		keysToEncrypt = append(keysToEncrypt, sec.Key)
	}
	encrypted, err := encrypt(b, recipients, keysToEncrypt)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(encrypted)
	return buf.Bytes(), err
}
