package dataservice

import (
	"testing"

	"github.com/GoSVI/impact-go/config"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func TestCypherRead(t *testing.T) {
	driver, err := neo4j.NewDriver(config.DbUrl, neo4j.BasicAuth(config.Username, config.Password, ""))
	if err != nil {
		t.Errorf("%v", err)
	}
	defer driver.Close()

	clients, err := ReadClient("github.com/gin-gonic/gin", "v1.8.2", driver)
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Logf("%v", clients)
}
