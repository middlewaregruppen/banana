package git

import (
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type Cloner struct {
	storer    storage.Storer
	cloneURL  string
	cloneTag  string
	clonePath string
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

// WithCloneURL returns a ClonerOpts with the provided clone url
func WithCloneURL(s string) ClonerOpts {
	return func(c *Cloner) {
		c.cloneURL = s
	}
}

// WithCloneTag returns a ClonerOpts with the provided clone tag
func WithCloneTag(s string) ClonerOpts {
	return func(c *Cloner) {
		c.cloneTag = s
	}
}

// Clone performs a git clone using into targetPath.
// If fsys is nil, an in-memory temporary filesystem will be used.
func (c *Cloner) Clone(fsys filesys.FileSystem) error {

	// Use HEAD if tag is not provided
	ref := plumbing.HEAD
	if len(c.cloneTag) > 0 {
		ref = plumbing.NewTagReferenceName(c.cloneTag)
	}

	// Setup storage for go-git
	fs := memfs.New()
	if c.storer == nil {
		c.storer = memory.NewStorage()
	}

	// Clone
	_, err := git.Clone(c.storer, fs, &git.CloneOptions{
		URL:           c.cloneURL,
		ReferenceName: ref,
		//SingleBranch: true, 	// SingleBranch: true doesn't work together with ReferenceName when remote repo uses main instead of master as branch name
		Depth: 1,
	})

	if err != nil {
		return err
	}

	// Walk over the given targetPath in a temporary mem fs and copy files/foldes
	// to the tiven filesystem
	err = util.Walk(fs, c.clonePath, func(rel string, info os.FileInfo, err error) error {
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

		// Create target folder structure
		err = fsys.MkdirAll(filepath.Dir(rel))
		if err != nil {
			return err
		}

		// Create dst file
		dst, err := fsys.Create(rel)
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

// NewCloner creates a new cloner using the opts provided
func NewCloner(s string, opts ...ClonerOpts) *Cloner {
	cloner := &Cloner{
		cloneURL:  s,
		storer:    memory.NewStorage(),
		clonePath: ".",
	}
	// Apply options
	for _, opt := range opts {
		opt(cloner)
	}
	return cloner
}
