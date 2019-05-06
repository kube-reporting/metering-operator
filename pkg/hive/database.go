package hive

import "github.com/operator-framework/operator-metering/pkg/db"

type DatabaseParameters struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func ExecuteCreateDatabase(queryer db.Queryer, params DatabaseParameters) error {
	query := generateCreateDatabaseSQL(params, false)
	_, err := queryer.Query(query)
	return err
}

func ExecuteDropDatabase(queryer db.Queryer, dbName string, ignoreNotExists, cascade bool) error {
	query := generateDropDatabaseSQL(dbName, ignoreNotExists, cascade)
	_, err := queryer.Query(query)
	return err
}
