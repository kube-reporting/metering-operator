package hive

import (
	"database/sql"
	"database/sql/driver"
)

const defaultBatchSize = int64(1000)

var (
	_ driver.Driver = &Driver{}
)

func init() {
	sql.Register("hive", &Driver{})
}

type Driver struct{}

func (*Driver) Open(dsn string) (driver.Conn, error) {
	return Open(dsn)
}
