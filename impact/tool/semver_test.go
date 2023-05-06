package tool

import (
	"fmt"
	"testing"
)

func TestIsMinorOrPatchUpdate(t *testing.T) {
	oldVersion, newVersion := "v1.2", "v1.2.2"
	if !isMinorOrPatchUpdate(oldVersion, newVersion) {
		t.Errorf("error: %v -> %v", oldVersion, newVersion)
	}
}

func TestParseServerUrl(t *testing.T) {
	path := "D:\\Code\\SemanticVersionStudy-data\\dataset\\semver\\semver_combine\\combine\\__github.com_0chain_zboxcli_.csv"
	record, err := ParseServerUrl(path)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	fmt.Printf("%v", record)
}
