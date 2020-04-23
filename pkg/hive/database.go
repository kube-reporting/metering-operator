package hive

import "github.com/kube-reporting/metering-operator/pkg/db"

type DatabaseParameters struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func ExecuteCreateDatabase(execer db.Execer, params DatabaseParameters) error {
	query := generateCreateDatabaseSQL(params, false)
	_, err := execer.Exec(query)
	return err
}

func ExecuteDropDatabase(execer db.Execer, dbName string, ignoreNotExists, cascade bool) error {
	query := generateDropDatabaseSQL(dbName, ignoreNotExists, cascade)
	_, err := execer.Exec(query)
	return err
}
