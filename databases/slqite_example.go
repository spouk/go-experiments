package databases

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jinzhu/gorm"
	"sync"
	"log"
	"time"
)

//---------------------------------------------------------------------------
//  type core
//---------------------------------------------------------------------------
type (
	SqliteStock struct {
		sync.WaitGroup
		sync.RWMutex
		DB *gorm.DB
	}
)

//---------------------------------------------------------------------------
//  type tables
//---------------------------------------------------------------------------
type (
	TestTable struct {
		ID       int64
		Email    string
		Password string
	}
)

//---------------------------------------------------------------------------
//  functions
//---------------------------------------------------------------------------
func NewSqliteStock(filenameDbs string) *SqliteStock {
	s := &SqliteStock{}
	db, err := s.openDbs(filenameDbs)
	if err != nil {
		log.Fatal(err)
	}
	s.DB = db
	return s
}
func (s *SqliteStock) openDbs(filenameDbs string, listTables []interface{}) (*gorm.DB, error) {
	//create/open database
	db, err := gorm.Open("sqlite", filenameDbs)
	if err != nil {
		return nil, err
	}
	//set WAL mode
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA cache_size = 2000;")
	db.Exec("PRAGMA default_cache_size = 2000;")
	db.Exec("PRAGMA page_size = 4096;")
	db.Exec("PRAGMA synchronous = NORMAL;")
	db.Exec("PRAGMA foreign_keys = on;")
	db.Exec("PRAGMA busy_timeout = 1;")
	//create tables if not  exists
	for _, x := range listTables {
		if !db.HasTable(x) {
			err := db.CreateTable(x).Error
			if err != nil {
				log.Print(err)
			}
		}
	}
	//return result
	return db, nil
}
func (s *SqliteStock) Run(countWriter, countReader int, timeTriger int) {
	var chanEnd = make(chan bool)
	for x := 0; x < countWriter; x ++ {
		s.Add(1)
		go s.writer(chanEnd)
	}
	for x := 0; x < countReader; x ++ {
		s.Add(1)
		go s.reader(chanEnd)
	}
	var timer = time.NewTimer(time.Second * time.Duration(timeTriger))
	<-timer.C

}
func (s *SqliteStock) writer(chanEnd chan bool) {

}
func (s *SqliteStock) reader(chanEnd chan bool) {}
