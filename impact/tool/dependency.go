package tool

import (
	"os"

	"github.com/gocarina/gocsv"
)

type Dependency struct {
	ModuleName           string `csv:"module_name"`
	GoVersion            string `csv:"go_version"`
	Version              string `csv:"tag"`
	Time                 string `csv:"tag_time"`
	RequireModuleName    string `csv:"require_module_name"`
	RequireModuleVersion string `csv:"require_module_version"`
	Indirect             bool   `csv:"require_module_indirect"`
}

func ParseDependencyFromFile(path string) ([]*Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dependencyList := []*Dependency{}
	if err = gocsv.UnmarshalFile(file, &dependencyList); err != nil {
		return nil, err
	}

	return dependencyList, nil
}
