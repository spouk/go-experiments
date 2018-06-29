package databases

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jinzhu/gorm"
	"sync"
	"log"
	"time"
	"github.com/spouk/gocheats/utils"
	"strings"
)

//---------------------------------------------------------------------------
//  type core
//---------------------------------------------------------------------------
type (
	SqliteStock struct {
		sync.WaitGroup
		sync.RWMutex
		DB         *gorm.DB
		Randomizer *utils.Randomizer
		Stock []*TestTable
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
func NewSqliteStock(filenameDbs string, listTables []interface{}) *SqliteStock {
	s := &SqliteStock{}
	db, err := s.openDbs(filenameDbs, listTables)
	if err != nil {
		log.Fatal(err)
	}
	s.DB = db
	s.Randomizer = utils.NewRandomize()
	return s
}
func (s *SqliteStock) openDbs(filenameDbs string, listTables []interface{}) (*gorm.DB, error) {
	//create/open database
	db, err := gorm.Open("sqlite3", filenameDbs)
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
func (s *SqliteStock) Run(countGenerator, countReader int, timeTriger int) {
	var chanEnd = make(chan bool)
	//run single writer
	s.Add(1)
	go s.writerDBSsingle(chanEnd)
	//run generators for stock
	for x := 0; x < countGenerator; x ++ {
		s.Add(1)
		go s.generatorStock(chanEnd, 1000)
	}
	//run reader from DBS
	for x := 0; x < countReader; x ++ {
		s.Add(1)
		go s.reader(chanEnd)
	}
	var timer = time.NewTimer(time.Second * time.Duration(timeTriger))
	<-timer.C
	log.Print("CLOSE CHANNEL FOR EXIT WORKERS\n")
	close(chanEnd)
	s.Wait()
	log.Print("all worker/reader end working")
}
func (s *SqliteStock) writer(chanEnd chan bool) {
	defer s.Done()
	for {
		select {
		case <-chanEnd:
			return
		default:
			var email = []string{s.Randomizer.RandomString(10), "@", s.Randomizer.RandomString(10), ".ru"}

			if err := s.DB.Create(&TestTable{Password: s.Randomizer.RandomString(10), Email: strings.Join(email, "")}); err != nil {
				log.Print(err)
			}
			time.Sleep(time.Second * 1)
		}
	}
}
func (s *SqliteStock) reader(chanEnd chan bool) {
	defer s.Done()
	for {
		select {
		case <-chanEnd:
			return
		default:
			var lists []TestTable
			if err := s.DB.Find(&lists).Error; err != nil {
				log.Print(err)
			} else {
				log.Printf("Count: %d\n", len(lists))
			}
			time.Sleep(time.Second * 1)
		}
	}
}
func (s *SqliteStock) generatorStock(chanEnd chan bool, countGenerator int) {
	defer s.Done()
	var trigger = 0
	for {
		select {
		case <- chanEnd:
			return
		default:
			if trigger < countGenerator {
				trigger ++
				s.Lock()
				var email = strings.Join([]string{s.Randomizer.RandomString(10), "@", s.Randomizer.RandomString(10), ".ru"}, "")
				s.Stock = append(s.Stock, &TestTable{Email:email, Password: s.Randomizer.RandomString(10)})
				s.Unlock()
				time.Sleep(time.Microsecond * 500)
			} else {
				log.Print("[generatorStock] END WORK\n")
				return
			}
		}
	}
}
func (s *SqliteStock) writerDBSsingle(chanEdn chan bool) {
	defer s.Done()
	for {
		select {
		case <- chanEdn:
			return
		default:
			if len(s.Stock) > 0 {
				//get element
				s.Lock()
				var element = s.Stock[0]
				s.Stock = append(s.Stock[:0], s.Stock[0+1:]...)
				s.Unlock()
				//write to dbs
				if err := s.DB.Create(element); err != nil {
					log.Print(err)
				}
			}
		}
	}
}