package tool

import (
	"fmt"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/hashicorp/go-version"
)

type Record struct {
	Name         string
	Url          string
	Incompatible map[string][]RecordDetail // the key represents version
}

type RecordDetail struct {
	Time          string
	PackagePath   string
	PackageName   string
	ChangeNode    string
	ChangeObject  string
	ChangeType    string
	ChangeMessage string
}

// should capital first letter of field becausing of gocsv
type checkResult struct {
	Url            string `csv:"server_url"`
	OldVersion     string `csv:"old_tag"`
	OldVersionTime string `csv:"olg_tag_time"`
	NewVersion     string `csv:"new_tag"`
	OldPkgPath     string `csv:"old_pkg_path"`
	OldPkgName     string `csv:"old_pkg_name"`
	ChangeNode     string `csv:"change_node"`
	ChangeObject   string `csv:"changedObject"`
	ChangeType     string `csv:"changedType"`
	ChangeMessage  string `csv:"changeMessage"`
}

func ParseServerUrl(path string) (*Record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	results := []*checkResult{}
	if err = gocsv.UnmarshalFile(file, &results); err != nil {
		return nil, err
	}
	record, err := extractRecord(results)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func extractRecord(results []*checkResult) (*Record, error) {
	recordDetail := make(map[string][]RecordDetail, 0)
	for _, result := range results {
		if !isMinorOrPatchUpdate(result.OldVersion, result.NewVersion) {
			continue
		}
		detail := RecordDetail{
			Time:          result.OldVersionTime,
			PackagePath:   result.OldPkgPath,
			PackageName:   result.OldPkgName,
			ChangeNode:    result.ChangeNode,
			ChangeType:    result.ChangeType,
			ChangeObject:  result.ChangeObject,
			ChangeMessage: result.ChangeMessage,
		}
		_, ok := recordDetail[result.OldVersion]
		if ok {
			recordDetail[result.OldVersion] = append(recordDetail[result.OldVersion], detail)
		} else {
			recordDetail[result.OldVersion] = []RecordDetail{detail}
		}
		fmt.Println(result)
	}
	name, url := parseNameAndUrl(results[0].Url)
	record := Record{
		Name:         name,
		Url:          url,
		Incompatible: recordDetail,
	}
	return &record, nil
}

func parseNameAndUrl(url string) (string, string) {
	url = url[:len(url)-4]
	urlSplit := strings.Split(url[:len(url)-4], "/")
	return strings.Join(urlSplit[3:5], "/"), url
}

func isMinorOrPatchUpdate(OldVersion, newVersion string) bool {
	v1, err := version.NewSemver(OldVersion)
	if err != nil {
		return false
	}
	v2, err := version.NewSemver(newVersion)
	if err != nil {
		return false
	}
	if v1.Prerelease() != "" || v1.Metadata() != "" {
		return false
	}
	if v2.Prerelease() != "" || v2.Metadata() != "" {
		return false
	}
	oldInfo := v1.Segments()
	newInfo := v2.Segments()
	if len(oldInfo) != len(newInfo) || (len(oldInfo) != 3 && len(oldInfo) != 2) {
		return false
	}
	if oldInfo[0] != newInfo[0] {
		return false
	}
	if oldInfo[1] < newInfo[1] || (len(oldInfo) == 3 && oldInfo[2] < newInfo[2]) {
		return true
	}
	panic("uncaught exception")
}
