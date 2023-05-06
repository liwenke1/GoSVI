package analyze

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoSVI/impact-go/config"
	"github.com/GoSVI/impact-go/store"
	"github.com/GoSVI/impact-go/tool"
	"golang.org/x/tools/go/packages"
)

func DetectImpact(item config.Item, defaultStorePath string) error {
	// create tmp dir to git clone repository
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// clone
	_, _, err = tool.RunCommand("git clone ", []string{item.Url, "."}, tmpDir)
	if err != nil {
		return err
	}

	for index, versionInfo := range item.Detail {
		fmt.Printf("remaining [%d/%d]\n", index+1, len(item.Detail))

		// checkout tag
		tool.RunCommand("git clean -fd && git checkout .", nil, tmpDir) // the change of tracking file will prevent tag's checkout
		_, _, err := tool.RunCommand("git checkout ", []string{versionInfo.Version}, tmpDir)
		if err != nil {
			log.Printf(": Checkout Tag Fail: %v -> %v", item.Url, versionInfo.Version)
			continue
		}

		// download mod deps
		if _, err := os.Stat(filepath.Join(tmpDir, "go.mod")); err != nil {
			log.Printf("repos version is not a go module: %v -> %v -> %v", item.Url, versionInfo.Version, strings.ReplaceAll(err.Error(), "\n", ""))
			continue
		}
		// if _, _, err := tool.RunCommand("go mod tidy", []string{}, tmpDir); err != nil {
		// 	os.Remove(fmt.Sprintf("%s/go.sum", tmpDir))
		// 	if _, _, err := tool.RunCommand("go mod tidy", []string{}, tmpDir); err != nil {
		// 		log.Printf("repos version can't download require deps: %v -> %v -> %v", item.Url, versionInfo.Version, strings.ReplaceAll(err.Error(), "\n", ""))
		// 		continue
		// 	}
		// }

		// parse identifier type and import pkg path
		pkgPathList := extractPkgPath(versionInfo.PkgInfo)
		if len(pkgPathList) == 0 {
			log.Printf("require third-party library pkg length is 0: %v -> %v", item.Url, versionInfo.Version)
			continue
		}
		identifierDetail, importDetail, err := parsePkgInfo(filepath.Join(tmpDir, versionInfo.Path), pkgPathList)
		if err != nil {
			log.Printf("load repos version error: %v -> %v -> %v", item.Url, versionInfo.Version, strings.ReplaceAll(err.Error(), "\n", ""))
			continue
		}

		if err = store.StoreImpactReport(store.RepositoryReport{
			ServerUrl:        url.URL{Scheme: item.Url},
			Version:          versionInfo.Version,
			IdentifierDetail: identifierDetail,
			ImportDetail:     importDetail,
		}, defaultStorePath); err != nil {
			log.Printf("store pkg info error: %v -> %v -> %v", item.Url, versionInfo.Version, strings.ReplaceAll(err.Error(), "\n", ""))
		}
	}
	return nil
}

func parsePkgInfo(dir string, pkgPathList []string) ([]store.IdentifierDetail, []store.ImportDetail, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		// Tests:      false,
		Dir: dir,
		// BuildFlags: []string{"-mod=mod"},
	}
	// go list 会自动忽略 gomod 目录下的 ventor 文件夹
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, nil, err
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) != 0 {
			return nil, nil, fmt.Errorf("package load not completed successfully, skip this tag")
		}
	}

	identifierDetail := []store.IdentifierDetail{}
	importDetail := []store.ImportDetail{}
	for _, pkg := range pkgs {
		for _, importPkg := range pkg.Imports {
			importDetail = append(importDetail, store.ImportDetail{
				PkgPath: importPkg.PkgPath,
				PkgName: importPkg.Name,
			})
		}
		identifierDetail = append(identifierDetail, parseIdentifierInfo(pkg.TypesInfo.Defs, pkg.Fset, "defs", pkgPathList)...)
		identifierDetail = append(identifierDetail, parseIdentifierInfo(pkg.TypesInfo.Uses, pkg.Fset, "uses", pkgPathList)...)

	}
	importDetail = filterDuplicatePkgPath(importDetail)
	return identifierDetail, importDetail, nil
}

func parseIdentifierInfo(identifieList map[*ast.Ident]types.Object, pkgFset *token.FileSet, varType string, pkgPathList []string) []store.IdentifierDetail {
	identifierDetail := []store.IdentifierDetail{}
	for id, obj := range identifieList {
		if obj == nil {
			continue
		}
		// check if identifier is come from third-party library
		if !contain(obj.String(), pkgPathList) {
			continue
		}
		objPkg := ""
		if obj.Pkg() != nil {
			objPkg = obj.Pkg().String()
		}
		identifierDetail = append(identifierDetail, store.IdentifierDetail{
			VarType:     varType,
			IdName:      id.Name,
			IdPos:       extractPosition(pkgFset.Position(id.Pos()).String()),
			ObjName:     obj.Name(),
			ObjPos:      pkgFset.Position(obj.Pos()).String(),
			ObjPkg:      objPkg,
			ObjType:     obj.Type().String(),
			ObjExported: obj.Exported(),
			ObjId:       obj.Id(),
			ObjString:   obj.String(),
		})
	}
	return identifierDetail
}

func contain(objString string, pkgInfo []string) bool {
	// need check the type of obj if exist in pkgInfo
	for _, pkgPath := range pkgInfo {
		if strings.Contains(objString, pkgPath) {
			return true
		}
	}
	return false
}

func extractPkgPath(pkgInfo []string) []string {
	pkgPathList := []string{}
	for _, info := range pkgInfo {
		path := strings.Split(info, "@@@")[0]
		pkgPathList = append(pkgPathList, path)
	}
	return pkgPathList
}

func extractPosition(absFilepath string) string {
	return strings.SplitN(absFilepath, "/", 4)[3]
}

func filterDuplicatePkgPath(importDetailList []store.ImportDetail) []store.ImportDetail {
	singlePkg := make(map[string]store.ImportDetail)
	for _, pkgDetail := range importDetailList {
		if _, exist := singlePkg[pkgDetail.PkgPath]; !exist {
			singlePkg[pkgDetail.PkgPath] = pkgDetail
		}
	}
	duplicateImportDetailList := []store.ImportDetail{}
	for _, pkgDetail := range singlePkg {
		duplicateImportDetailList = append(duplicateImportDetailList, pkgDetail)
	}
	return duplicateImportDetailList
}
