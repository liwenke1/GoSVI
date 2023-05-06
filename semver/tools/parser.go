// Package tools has two usages
// one is get all tags of given remote url, this usage is not exported
// another one is get all *types.packages between two nearby tags
package tools

import (
	"errors"
	"fmt"
	"go/types"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoSVI/semver/store"
	"github.com/GoSVI/semver/utils/cmd"

	"golang.org/x/mod/semver"
	"golang.org/x/tools/go/packages"
)

type pkgPair struct {
	serverUrl                url.URL
	oldTag, newTag           string
	oldTagTime, newTagTime   string
	oldPkgs, newPkgs         map[string]*types.Package
	oldPkgsPath, newPkgsPath map[string]string
}

// getPkgPair returns all *types.Packages between any two nearby tags with given repository
// address is url
func getPkgPair(url url.URL, storePath string, whiteList []string) error {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// clone
	_, _, err = cmd.RunCommand("git clone ", []string{url.Scheme, "."}, tmpDir)
	if err != nil {
		return err
	}

	// get tags
	f, _, err := cmd.RunCommand("git tag --sort=v:refname", nil, tmpDir)
	if err != nil {
		return err
	}
	ts := strings.Split(f.String(), "\n")
	if len(ts) < 2 {
		return errors.New("the tags of repo is less than two")
	}

	// detect each tag
	var newTagCommitTime string
	newPkgs := make(map[string]*types.Package)
	newPkgsPathMap := make(map[string]string)
	for i := 0; i <= len(ts)-1; i++ {
		fmt.Printf("remaining [%d/%d]\n", i+1, len(ts))
		t := ts[i]

		t1 := t
		if strings.HasPrefix(t1, "v") {
			t1 = t[1:]
		}
		if !semver.IsValid(t1) {
			continue
		}

		tagCommitTime, pkgs, pkgsPathMap, err := getPackagesOfTag(tmpDir, t)
		if err != nil {
			log.Printf(": getPackagesOfTag err: %s, url: %v, tag: %s\n", strings.ReplaceAll(err.Error(), "\n", ""), url.Scheme, t)
			continue
		}

		oldPkgs := newPkgs
		newPkgs = pkgs
		oldTagCommitTime := newTagCommitTime
		newTagCommitTime = tagCommitTime
		oldPkgsPathMap := newPkgsPathMap
		newPkgsPathMap = pkgsPathMap

		if len(oldPkgs) > 0 {
			if ret, err := diffPkg(pkgPair{
				serverUrl:   url,
				oldTag:      ts[i-1],
				newTag:      t,
				oldTagTime:  oldTagCommitTime,
				newTagTime:  newTagCommitTime,
				oldPkgs:     oldPkgs,
				newPkgs:     newPkgs,
				oldPkgsPath: oldPkgsPathMap,
				newPkgsPath: newPkgsPathMap,
			}, whiteList); err == nil {
				if err := store.StoreReports([]store.Report{ret}, filepath.Join(storePath, "semver")); err != nil {
					return err
				}
			}
		}
		log.Printf(": parse tag success: %v -> %v\n", url.Scheme, t)
	}
	return nil
}

// getPackagesOfVersion returns the *ast.packages between two given tags
func getPackagesOfTag(tmpDir, tag string) (string, map[string]*types.Package, map[string]string, error) {
	// checkout
	cmd.RunCommand("git clean -fd && git checkout .", nil, tmpDir) // the change of tracking file will prevent tag's checkout
	_, _, err := cmd.RunCommand("git checkout ", []string{tag}, tmpDir)
	if err != nil {
		return "", nil, nil, err
	}

	// download mod deps
	if _, err := os.Stat(filepath.Join(tmpDir, "go.mod")); err != nil {
		return "", nil, nil, fmt.Errorf("not a go module: %v", err)
	}
	if _, _, err := cmd.RunCommand("go mod tidy", []string{}, tmpDir); err != nil {
		os.Remove(fmt.Sprintf("%s/go.sum", tmpDir))
		if _, _, err := cmd.RunCommand("go mod tidy", []string{}, tmpDir); err != nil {
			return "", nil, nil, fmt.Errorf("can't download require deps: %v", err)
		}
	}

	// get commit time of tag
	tagCommitTime, _, _ := cmd.RunCommand("git log -1 --format=%aI ", []string{tag}, tmpDir)

	typePkgs, pkgPaths, err := load(tmpDir)

	// load packages
	return strings.Split(tagCommitTime.String(), "\n")[0], typePkgs, pkgPaths, err
}

func load(dir string) (map[string]*types.Package, map[string]string, error) {
	cfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedImports | packages.NeedDeps | packages.NeedTypes,
		Tests:      false,
		Dir:        dir,
		BuildFlags: []string{"-mod=mod"},
	}
	// go list 会自动忽略 gomod 目录下的 ventor 文件夹
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, nil, err
	}
	return convert(pkgs)
}

func convert(pkgs []*packages.Package) (map[string]*types.Package, map[string]string, error) {
	typePkgs := make(map[string]*types.Package, 0)
	pkgPaths := make(map[string]string, 0)
	for _, pkg := range pkgs {
		if len(pkg.Errors) != 0 {
			return nil, nil, fmt.Errorf("package load not completed successfully, skip this tag")
		}
		typePkgs[pkg.Name] = pkg.Types
		pkgPaths[pkg.Name] = pkg.PkgPath
	}
	return typePkgs, pkgPaths, nil
}
