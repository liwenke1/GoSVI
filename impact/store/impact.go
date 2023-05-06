package store

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type RepositoryReport struct {
	ServerUrl        url.URL
	Version          string
	IdentifierDetail []IdentifierDetail
	ImportDetail     []ImportDetail
}

type IdentifierDetail struct {
	VarType     string
	IdName      string
	IdPos       string
	ObjName     string
	ObjPos      string
	ObjPkg      string
	ObjType     string
	ObjExported bool
	ObjId       string
	ObjString   string
}

type ImportDetail struct {
	PkgPath string
	PkgName string
}

// StoreReports store the report data to a .csv format file
// store path: ./report/
// store file name: server_name.csv
func StoreImpactReport(report RepositoryReport, storePath string) error {
	pkgInfoPath := filepath.Join(storePath, "pkgInfo")
	err := storePkgDetailToFile(report, pkgInfoPath)
	if err != nil {
		return err
	}
	importInfoPath := filepath.Join(storePath, "importPkg")
	err = storeImportDetailToFile(report, importInfoPath)
	if err != nil {
		return err
	}
	return nil
}

func storePkgDetailToFile(report RepositoryReport, storePath string) error {
	var csvMsg [][]string
	serverName := getServerNameFromUrl(report.ServerUrl.Scheme)
	for _, identifierInfo := range report.IdentifierDetail {
		m := []string{
			report.ServerUrl.Scheme,
			identifierInfo.VarType,
			identifierInfo.IdName,
			identifierInfo.IdPos,
			identifierInfo.ObjName,
			identifierInfo.ObjPos,
			identifierInfo.ObjPkg,
			identifierInfo.ObjType,
			strconv.FormatBool(identifierInfo.ObjExported),
			identifierInfo.ObjId,
			identifierInfo.ObjString,
		}
		csvMsg = append(csvMsg, m)

	}
	if len(csvMsg) == 0 {
		return nil
	}
	dir := fmt.Sprintf("%s/%s_@@@%s.csv", storePath, serverName, strings.ReplaceAll(report.Version, "/", "_"))
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(storePath, 0777)
		csvMsg = append([][]string{
			{
				"server_url",
				"var_type",
				"id_name",
				"id_pos",
				"obj_name",
				"obj_pos",
				"obj_pkg",
				"obj_type",
				"obj_exported",
				"obj_id",
				"obj_string",
			},
		}, csvMsg...)
	}
	file, err := os.OpenFile(dir, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err = writer.WriteAll(csvMsg); err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func storeImportDetailToFile(report RepositoryReport, storePath string) error {
	var csvMsg [][]string
	serverName := getServerNameFromUrl(report.ServerUrl.Scheme)
	for _, importInfo := range report.ImportDetail {
		m := []string{
			report.ServerUrl.Scheme,
			importInfo.PkgPath,
			importInfo.PkgName,
		}
		csvMsg = append(csvMsg, m)

	}
	if len(csvMsg) == 0 {
		return nil
	}
	dir := fmt.Sprintf("%s/%s_@@@%s.csv", storePath, serverName, strings.ReplaceAll(report.Version, "/", "_"))
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(storePath, 0777)
		csvMsg = append([][]string{
			{
				"server_url",
				"pkg_path",
				"pkg_name",
			},
		}, csvMsg...)
	}
	file, err := os.OpenFile(dir, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err = writer.WriteAll(csvMsg); err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func getServerNameFromUrl(serverUrl string) string {
	var serverName string
	idx := strings.Index(serverUrl, ":")
	if idx != -1 {
		serverName = serverUrl[idx+1:]
	}
	serverName = strings.TrimSuffix(serverName, ".git")
	serverName = strings.ReplaceAll(serverName, "/", "_")
	return serverName
}
