package main

import "github.com/spouk/go-experiments/databases"


func main() {
	var dbsfile  = "testdbs.dbs"
	var listTables =[]interface{}{databases.TestTable{}}

	db := databases.NewSqliteStock(dbsfile, listTables)
	db.Run(10,10, 5)
}
