package git

import (
	"testing"

	"github.com/middlewaregruppen/banana/api/types"
	"github.com/middlewaregruppen/banana/pkg/module"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var tmpfs = filesys.MakeFsInMemory()

func makeModule(mod types.Module) module.Module {
	return module.NewKustomizeModule(tmpfs, mod, "testprefix")
}

func TestModuleVersion(t *testing.T) {
	tests := []struct {
		name  string
		want  string
		input types.Module
	}{
		{
			"no version",
			"HEAD",
			types.Module{},
		},
		{
			"tag version",
			"refs/tags/v1.0.0",
			types.Module{
				Version: "v1.0.0",
			},
		},
		{
			"branch ref",
			"refs/heads/test-branch-name",
			types.Module{
				Ref: "refs/heads/test-branch-name",
			},
		},
		{
			"both ref and version",
			"refs/heads/test-branch-name",
			types.Module{
				Ref:     "refs/heads/test-branch-name",
				Version: "v1.0.0",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cloner := NewCloner(makeModule(test.input))
			assert.Equal(t, test.want, cloner.GetRef().String())
		})
	}
}
