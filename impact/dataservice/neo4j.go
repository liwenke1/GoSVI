package dataservice

import (
	"log"

	"github.com/GoSVI/impact-go/tool"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Repository struct {
	Name    string
	Version string
	Url     string
}

func StoreToNeo4j(task string, driver neo4j.Driver) error {
	dependencyList, err := tool.ParseDependencyFromFile(task)
	if err != nil {
		return err
	}

	if err = cypherWrite(dependencyList, driver); err != nil {
		return err
	}

	return nil
}

func cypherWrite(dependencyList []*tool.Dependency, driver neo4j.Driver) error {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	for _, dependency := range dependencyList {
		if dependency.Version == "" {
			continue
		}
		_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			_, err := tx.Run("MERGE (n1 {Name:$Name1, Version:$Version1}) MERGE (n2 {Name:$Name2, Version:$Version2}) MERGE (n1)-[r:import {Indirect:toBoolean($Indirect)}]->(n2)",
				map[string]interface{}{"Name1": dependency.ModuleName, "Version1": dependency.Version, "Name2": dependency.RequireModuleName, "Version2": dependency.RequireModuleVersion, "Indirect": dependency.Indirect})
			if err != nil {
				return nil, err
			}
			return nil, err
		})
		if err != nil {
			log.Println("write to DB with error: ", err)
			return err
		}
	}
	return nil
}

func ReadClient(url, version string, driver neo4j.Driver) ([]Repository, error) {
	clients, err := cypherRead(url, version, driver)
	if err != nil {
		return nil, err
	}
	return clients, nil
}

func cypherRead(url, version string, driver neo4j.Driver) ([]Repository, error) {
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()
	clients := make([]Repository, 0)

	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run("match(n1)-[r]->(n2) where n1.Url<>'' and n2.Url=$Url and n2.Version=$Version and r.Indirct=false return n1",
			map[string]interface{}{"Url": url, "Version": version})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			record := result.Record()
			if value, ok := record.Get("n1"); ok {
				node := value.(neo4j.Node)
				clients = append(clients, Repository{
					Name:    node.Props["Name"].(string),
					Version: node.Props["Version"].(string),
					Url:     node.Props["Url"].(string),
				})
			}
		}
		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return clients, nil
}
