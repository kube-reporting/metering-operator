# Go-Hive-Driver

A hive driver for Go's [database/sql](https://golang.org/pkg/database/sql/) package

## Features
* Support hive thrift server 2
* Support SASL
* Implement database/sql package

## Requirements
* Go 1.10+
* Hive thrift server 2

## Usage
```go
package main

import (
	"database/sql"
	_ "github.com/taozle/go-hive-driver"
)

func main()  {
        db, err := sql.Open("hive", "hive://user:password@host:port?auth=sasl&batch=500")
}
```