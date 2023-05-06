package tool

import "testing"

func TestParseDependencyFromFile(t *testing.T) {
	path := "D:\\Code\\SemanticVersionStudy-data\\dataset\\dependency\\combine\\__github.com_0chain_zboxcli_.csv"
	dependencyList, err := ParseDependencyFromFile(path)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	t.Logf("%v", dependencyList)
}
