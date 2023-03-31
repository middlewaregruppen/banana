package module

import (
	"fmt"
	"html/template"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/amimof/kmaint/api/types"
)

var (
	TemplateSuffix = ".tmpl"
)

type Module interface {
	Build() error
}

type builtinModule struct {
	opts types.ModuleOpts
}

type remoteModule struct {
	opts types.ModuleOpts
}

func (m builtinModule) Build() error {
	// Create the folder structure
	modulePath := fmt.Sprintf("%s/%s", "modules", m.opts.ModuleName())
	srcPath := fmt.Sprintf("%s/%s", "src", m.opts.ModuleName())

	// First we check if module exists
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return err
	}

	// Create folder structure
	err := os.MkdirAll(srcPath, os.ModePerm)
	if err != nil {
		return err
	}

	// Walk the folder structure and attempt to find template files
	err = filepath.Walk(modulePath, func(rel string, info os.FileInfo, err error) error {

		// Ignore directories
		if info.IsDir() {
			return err
		}

		// Replace leading path "modules/" with "src/" and create folder structure within it
		dirs := strings.Split(rel, string(os.PathSeparator))
		dirs[0] = "src"
		dstName := path.Join(dirs...)
		dstDir := filepath.Dir(dstName)
		err = os.MkdirAll(dstDir, os.ModePerm)
		if err != nil {
			return err
		}

		// Remove template suffix from file name
		if strings.HasSuffix(dstName, TemplateSuffix) {
			dstName = path.Join(dstDir, info.Name()[:len(info.Name())-len(TemplateSuffix)])
		}

		// Stat file and check if it's a regular file
		srcStat, err := os.Stat(rel)
		if err != nil {
			return err
		}
		if !srcStat.Mode().IsRegular() {
			return err
		}

		// Open file for reading
		srcFile, err := os.Open(rel)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Create destination file which we will be writing to
		dstFile, err := os.Create(dstName)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// We only care about files ending in .tmpl when templating
		if !strings.HasSuffix(rel, TemplateSuffix) {
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}
		}

		// Parse template injecting variables into it
		tmpl, err := template.ParseFiles(rel)
		if err != nil {
			return err
		}

		// Write output to file
		err = tmpl.Execute(dstFile, m.opts)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (m remoteModule) Build() error {
	return nil
}

func LoadBuiltin(opts types.ModuleOpts) Module {
	return &builtinModule{opts: opts}
}

func LoadRemote(opts types.ModuleOpts, u *url.URL) Module {
	return &remoteModule{opts: opts}
}

func Load(m types.ModuleOpts) Module {
	if u, err := url.ParseRequestURI(m.ModuleName()); err == nil {
		return LoadRemote(m, u)
	}
	// If all above fail then it must be a builtin module
	return LoadBuiltin(m)
}

func WithParentOpts(km *types.KmaintFile, m types.Module) types.ModuleOpts {
	newModule := m
	newModule.Opts["Name"] = km.Name
	newModule.Opts["Version"] = km.Version
	newModule.Opts["Metadata"] = km.MetaData
	newModule.Opts["Clusters"] = km.Clusters
	newModule.Opts["Module"] = m
	return newModule.Opts
}
