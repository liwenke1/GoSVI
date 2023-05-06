package store

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/GoSVI/semver/apidiff"
)

// Report record the diff between tags
type Report struct {
	ServerUrl      url.URL
	OldVersion     string
	OldVersionTime string
	NewVersion     string
	NewVersionTime string
	Detail         []PkgDetail
}

// PkgDetail record the diff detail message between packages
type PkgDetail struct {
	OldPkgName string
	NewPkgName string
	OldPkgPath string
	NewPkgPath string
	Result     apidiff.Report
}

// StoreReports store the report data to a .csv format file
// store path: ./report/
// store file name: server_name.csv
func StoreReports(reports []Report, storePath string) error {
	for _, reportItem := range reports {
		err := storePkgDetailToFile(reportItem, storePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func storePkgDetailToFile(reportMsg Report, storePath string) error {
	var csvMsg [][]string
	serverName := getServerNameFromUrl(reportMsg.ServerUrl.Scheme)
	for _, pkgDetailItem := range reportMsg.Detail {
		for _, change := range pkgDetailItem.Result.Changes {
			if strings.Contains(change.Message, "from unknown to unknown") {
				continue
			}
			if change.Compatible {
				continue
			}
			changeNode := change.Node
			changedObject := change.ChangedObject
			changedType := change.ChangedType
			changeMsg := change.Message
			m := []string{
				serverName,
				reportMsg.ServerUrl.Scheme,
				reportMsg.OldVersion,
				reportMsg.OldVersionTime,
				reportMsg.NewVersion,
				reportMsg.NewVersionTime,
				pkgDetailItem.OldPkgName,
				pkgDetailItem.NewPkgName,
				pkgDetailItem.OldPkgPath,
				pkgDetailItem.NewPkgPath,
				changeNode,
				changedObject,
				changedType,
				changeMsg,
			}
			csvMsg = append(csvMsg, m)
		}
	}
	if len(csvMsg) == 0 {
		return nil
	}
	dir := fmt.Sprintf("%s/%s_.csv", storePath, serverName)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(storePath, 0777)
		csvMsg = append([][]string{
			{
				"server_name",
				"server_url",
				"old_tag",
				"olg_tag_time",
				"new_tag",
				"new_tag_time",
				"old_pkg_name",
				"new_pkg_name",
				"old_pkg_path",
				"new_pkg_path",
				"change_node",
				"changedObject",
				"changedType",
				"changeMessage",
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
