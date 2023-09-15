package module

import (
	"bytes"
	"testing"

	"github.com/middlewaregruppen/banana/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func newModule(mod types.Module) Module {
	return NewKustomizeModule(filesys.MakeFsInMemory(), mod, "src")
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
				Name: "monitoring/grafana",
				Host: &types.Host{
					Prefix: "infra",
				},
			},
			"asd",
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
			if buf.String() != tt.want {
				t.Fatalf("got %s but wanted %s", buf.String(), tt.want)
			}
		})
	}

}
