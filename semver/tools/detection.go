// Package tools has two usages
// one is detect packages changed details between two nearby tags
// another is detect packages changed details between two given tags
package tools

import (
	"fmt"
	"go/types"
	"log"
	"net/url"
	"strings"

	"github.com/GoSVI/semver/apidiff"
	"github.com/GoSVI/semver/store"
)

// DetectSemVer detect imcompatiable changes of
// all neary tags of a given repo url
func DetectSemVer(serverUrl url.URL, storePath string, whiteList []string) error {
	return getPkgPair(serverUrl, storePath, whiteList)
}

func diffPkg(pair pkgPair, whiteList []string) (store.Report, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf(": Skip Panic - %s %s -> %s\n", &pair.serverUrl, pair.oldTag, pair.newTag)
		}
	}()
	oldPkgs, newPkgs := pair.oldPkgs, pair.newPkgs
	oldPkgsPath, newPkgsPath := pair.oldPkgsPath, pair.newPkgsPath
	ret := store.Report{
		ServerUrl:      pair.serverUrl,
		OldVersion:     pair.oldTag,
		OldVersionTime: pair.oldTagTime,
		NewVersion:     pair.newTag,
		NewVersionTime: pair.newTagTime,
	}
	filtWhiteList(oldPkgs, oldPkgsPath, whiteList)
	filtWhiteList(newPkgs, newPkgsPath, whiteList)
	for path, oldPkg := range oldPkgs {
		// old and new matchresult
		if newPkg, ok := newPkgs[path]; ok {
			result := apidiff.Changes(oldPkg, newPkg)
			if len(result.Changes) != 0 {
				ret.Detail = append(ret.Detail, store.PkgDetail{
					OldPkgName: oldPkg.Name(),
					NewPkgName: newPkg.Name(),
					OldPkgPath: oldPkgsPath[oldPkg.Name()],
					NewPkgPath: newPkgsPath[newPkg.Name()],
					Result:     result,
				})
			}
		} else {
			// old has while new does not has
			result := apidiff.Change{Message: fmt.Sprintf("%s: packages deleted", oldPkg.Name()), Compatible: false}
			ret.Detail = append(ret.Detail, store.PkgDetail{
				OldPkgName: oldPkg.Name(),
				OldPkgPath: oldPkgsPath[oldPkg.Name()],
				Result: apidiff.Report{
					Changes: []apidiff.Change{result},
				},
			})
		}
	}
	for path, newPkg := range newPkgs {
		if _, ok := oldPkgs[path]; !ok {
			result := apidiff.Change{Message: fmt.Sprintf("%s: packages new added", newPkg.Name()), Compatible: true}
			ret.Detail = append(ret.Detail, store.PkgDetail{
				NewPkgName: newPkg.Name(),
				NewPkgPath: newPkgsPath[newPkg.Name()],
				Result: apidiff.Report{
					Changes: []apidiff.Change{result},
				},
			})
		}
	}
	return ret, nil
}

func filtWhiteList(pkgs map[string]*types.Package, paths map[string]string, whiteList []string) {
	for k := range pkgs {
		if inWhiteList(paths[k], whiteList) {
			delete(pkgs, k)
		}
	}
}

func inWhiteList(path string, whiteList []string) bool {
	for _, e := range whiteList {
		if strings.Contains(path, "/"+e) || strings.HasPrefix(path, e+"/") {
			return true
		}
	}
	return false
}
