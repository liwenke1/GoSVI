package analyze

import (
	"fmt"
	"testing"

	"github.com/GoSVI/impact-go/store"
)

func TestExtractPkgPath(t *testing.T) {
	pkgInfo := []string{
		"github.com/golang-migrate/migrate/v4/database/cockroachdb@@@cockroachdb",
		"github.com/golang-migrate/migrate/v4/database/firebird@@@firebird",
		"github.com/golang-migrate/migrate/v4/database/mongodb@@@mongodb",
		"github.com/golang-migrate/migrate/v4/database/mysql@@@mysql",
		"github.com/golang-migrate/migrate/v4/database/ql@@@ql",
		"github.com/golang-migrate/migrate/v4/database/postgres@@@postgres",
		"github.com/golang-migrate/migrate/v4/database/cassandra@@@cassandra",
		"github.com/golang-migrate/migrate/v4/database/sqlcipher@@@sqlcipher",
		"github.com/swaggo/swag@@@swag",
		"github.com/golang-migrate/migrate/v4/database/sqlite3@@@sqlite3",
		"github.com/golang-migrate/migrate/v4/database/sqlite@@@sqlite",
		"github.com/golang-migrate/migrate/v4/database/sqlserver@@@sqlserver",
		"github.com/swaggo/swag/gen@@@gen",
		"github.com/golang-migrate/migrate/v4/database/redshift@@@redshift",
		"github.com/golang-migrate/migrate/v4/database/pgx@@@pgx",
		"github.com/golang-migrate/migrate/v4/database/clickhouse@@@clickhouse",
		"github.com/golang-migrate/migrate/v4/database/snowflake@@@snowflake",
	}
	destPkgPathList := []string{
		"github.com/golang-migrate/migrate/v4/database/cockroachdb",
		"github.com/golang-migrate/migrate/v4/database/firebird",
		"github.com/golang-migrate/migrate/v4/database/mongodb",
		"github.com/golang-migrate/migrate/v4/database/mysql",
		"github.com/golang-migrate/migrate/v4/database/ql",
		"github.com/golang-migrate/migrate/v4/database/postgres",
		"github.com/golang-migrate/migrate/v4/database/cassandra",
		"github.com/golang-migrate/migrate/v4/database/sqlcipher",
		"github.com/swaggo/swag",
		"github.com/golang-migrate/migrate/v4/database/sqlite3",
		"github.com/golang-migrate/migrate/v4/database/sqlite",
		"github.com/golang-migrate/migrate/v4/database/sqlserver",
		"github.com/swaggo/swag/gen",
		"github.com/golang-migrate/migrate/v4/database/redshift",
		"github.com/golang-migrate/migrate/v4/database/pgx",
		"github.com/golang-migrate/migrate/v4/database/clickhouse",
		"github.com/golang-migrate/migrate/v4/database/snowflake",
	}
	pkgPathList := extractPkgPath(pkgInfo)
	if len(pkgPathList) != len(pkgInfo) {
		t.Errorf("parse length error: %v -> %v", len(pkgInfo), len(pkgPathList))
	}
	for i := 0; i < len(pkgPathList); i++ {
		if pkgPathList[i] != destPkgPathList[i] {
			t.Errorf("parse path error: %v -> %v", pkgInfo[i], pkgPathList[i])
		}
	}
}

func TestFilterDuplicatePkgPath(t *testing.T) {
	importDetailList := []store.ImportDetail{
		{PkgPath: "ddd", PkgName: "dd.name"},
		{PkgPath: "ddd", PkgName: "dd.name"},
		{PkgPath: "aaa", PkgName: "aa.name"},
	}
	fmt.Println(importDetailList)
	fmt.Println(filterDuplicatePkgPath(importDetailList))
}
