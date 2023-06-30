package git

import (
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

func Clone(fsys filesys.FileSystem, targetURL, targetPath string) error {
	fs := memfs.New()
	mem := memory.NewStorage()
	_, err := git.Clone(mem, fs, &git.CloneOptions{
		URL: targetURL,
	})
	if err != nil {
		return err
	}

	// Walk over the given targetPath in a temporary mem fs and copy files/foldes
	// to the tiven filesystem
	err = util.Walk(fs, targetPath, func(rel string, info os.FileInfo, err error) error {
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
