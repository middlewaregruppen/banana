package module

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/middlewaregruppen/banana/api/types"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var testfs filesys.FileSystem

func init() {
	testfs = filesys.MakeFsInMemory()

	// Simple Ingress
	ing := []byte(ingressData)

	err := makeSingleModule(ing)
	if err != nil {
		panic(err)
	}
}

func makeSingleModule(data []byte) error {
	// The kustomization
	kust := []byte(kustomizationData)

	// Create module folder structure
	rootdir := "test-namespace/test-module"
	err := testfs.MkdirAll(rootdir)
	if err != nil {
		return err
	}

	// Create kustomization file in tmp fs
	kustf, err := testfs.Create(fmt.Sprintf("%s/%s", rootdir, "kustomization.yaml"))
	if err != nil {
		return err
	}
	defer kustf.Close()
	_, err = kustf.Write(kust)
	if err != nil {
		return err
	}

	// Create the resource
	resf, err := testfs.Create(fmt.Sprintf("%s/%s", rootdir, "resource.yaml"))
	if err != nil {
		return err
	}
	defer resf.Close()
	_, err = resf.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func newModule(mod types.Module) Module {
	return NewKustomizeModule(testfs, mod, "src")
}

func TestKustomizeModuleBuild_Ingresses(t *testing.T) {
	var tests = []struct {
		name  string
		input types.Module
		want  string
	}{
		{
			"ingress with prefix",
			types.Module{
				Name: "test-namespace/test-module",
				Host: &types.Host{
					Prefix: "infra",
				},
			},
			`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress
  namespace: test-namespace
spec:
  ingressClassName: nginx
  rules:
  - host: infra-test-ingress
    http:
      paths:
      - backend:
          service:
            name: test-service
            port:
              number: 80
        path: /
        pathType: Prefix
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			m := newModule(tt.input)
			err := m.Build(&buf)
			if err != nil {
				t.Fatal(err)
			}
			if assert.Equal(t, tt.want, buf.String()) {
				t.Fatalf("got %s but wanted %s", buf.String(), tt.want)
			}
		})
	}

}
