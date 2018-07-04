package main

import "github.com/spouk/go-experiments/databases"
//import (
//	"github.com/spouk/gocheats/utils"
//	"fmt"
//
//)


func main() {
	//r := utils.NewRandomize()
	//for x:=0; x < 100; x++ {
	//	fmt.Printf(" %v \n", r.RandomStringChoice(20, utils.LasciiLetters))
	//}

	var dbsfile  = "testdbs.dbs"
	var listTables =[]interface{}{databases.TestTable{}}

	db := databases.NewSqliteStock(dbsfile, listTables)
	db.RunerGeneratorWorkerDBS(10)
	//db.Run(10,10, 5)
}
