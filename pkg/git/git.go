package git

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/middlewaregruppen/banana/pkg/module"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type Cloner struct {
	// Repository storer
	storer storage.Storer

	// The git URL to clone
	cloneURL string

	// The git ref to clone
	cloneRef plumbing.ReferenceName

	// The path to which clone into
	clonePath string

	// The sub directory in the repo to clone
	cloneSubdir string
}

type ClonerOpts func(c *Cloner)

// WithStorer returns a ClonerOpts with the provided storer
func WithStorer(s storage.Storer) ClonerOpts {
	return func(c *Cloner) {
		c.storer = s
	}
}

// WithTargetPath returns a ClonerOpts with the provided target path
func WithTargetPath(s string) ClonerOpts {
	return func(c *Cloner) {
		c.clonePath = s
	}
}

// // WithCloneURL returns a ClonerOpts with the provided clone url
// func WithCloneURL(s string) ClonerOpts {
// 	return func(c *Cloner) {
// 		c.cloneURL = s
// 	}
// }

// // WithCloneRef returns a ClonerOpts with the provided clone ref
// func WithCloneRef(s string) ClonerOpts {
// 	return func(c *Cloner) {
// 		c.cloneRef = s
// 	}
// }

func WithCloneSubDir(s string) ClonerOpts {
	return func(c *Cloner) {
		c.cloneSubdir = s
	}
}

// Clone performs a git clone using into targetPath.
// If fsys is nil, an in-memory temporary filesystem will be used.
func (c *Cloner) Clone(fsys filesys.FileSystem) error {

	// Setup storage for go-git
	fs := memfs.New()
	if c.storer == nil {
		c.storer = memory.NewStorage()
	}

	// Clone
	_, err := git.Clone(c.storer, fs, &git.CloneOptions{
		URL:           c.cloneURL,
		ReferenceName: c.cloneRef,
		//SingleBranch: true, 	// SingleBranch: true doesn't work together with ReferenceName when remote repo uses main instead of master as branch name
		Depth: 1,
	})

	if err != nil {
		return err
	}

	// Walk over the given targetPath in a temporary mem fs and copy files/folders
	// to the tiven filesystem
	err = util.Walk(fs, c.cloneSubdir, func(rel string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}

		// Open source file for reading, close when done
		src, err := fs.Open(rel)
		if err != nil {
			return err
		}
		defer src.Close()

		// Stat file and check if it's a regular file
		srcStat, err := fs.Stat(rel)
		if err != nil {
			return err
		}

		// Ensure file is regular
		if !srcStat.Mode().IsRegular() {
			return err
		}

		dstRel := fmt.Sprintf("%s/%s", c.clonePath, rel)

		// Create target folder structure
		err = fsys.MkdirAll(filepath.Dir(dstRel))
		if err != nil {
			return err
		}

		// Create dst file
		dst, err := fsys.Create(dstRel)
		if err != nil {
			return err
		}
		defer dst.Close()

		// Begin copy source to destination file
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Cloner) GetRef() plumbing.ReferenceName {
	return c.cloneRef
}

func NewCloner(mod module.Module, opts ...ClonerOpts) *Cloner {
	cloneRef := plumbing.HEAD
	if len(mod.Version()) > 0 {
		cloneRef = plumbing.NewTagReferenceName(mod.Version())
	}
	if len(mod.Ref()) > 0 {
		cloneRef = plumbing.ReferenceName(mod.Ref())
	}
	cloner := &Cloner{
		cloneURL:    mod.URL(),
		cloneRef:    cloneRef,
		storer:      memory.NewStorage(),
		clonePath:   ".",
		cloneSubdir: ".",
	}
	// Apply options
	for _, opt := range opts {
		opt(cloner)
	}
	return cloner
}
