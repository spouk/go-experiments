package databases

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jinzhu/gorm"
	"sync"
	"log"
	"time"
	"github.com/spouk/gocheats/utils"
	"strings"
	"fmt"
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
		Stock      []*TestTable
		StockRead  []TestTable
		StockWrite []TestTable
		endChan    chan bool
	}
)

//---------------------------------------------------------------------------
//  type tables
//---------------------------------------------------------------------------
type (
	TestTable struct {
		ID       uint
		Email    string
		Password string
		ExHash   string
		Active   bool
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
	s.endChan = make(chan bool)
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

//---------------------------------------------------------------------------
//  пример работы при ситуации 1 генератор данных в базу данных
//
//---------------------------------------------------------------------------
func (s *SqliteStock) RunerGeneratorWorkerDBS(countWorker int) {
	//run generator
	s.Add(1)
	go s.generator()
	//run workerDBS
	s.Add(1)
	go s.workerDBS()
	//run workers
	for i := 0; i < countWorker; i ++ {
		s.Add(1)
		go s.workerExample(fmt.Sprintf("W#%d", i))
	}
	var triggerEnd = time.NewTimer(time.Second * 5)
	<-triggerEnd.C
	close(s.endChan)
	s.Wait()
	log.Println("THE END")
	for i, x := range s.StockWrite {
		fmt.Printf("[%d]::: %v\n", i, x)
	}
}
func (s *SqliteStock) generator() {
	log.Println("generator starting")
	defer func() {
		s.Done()
		log.Println("generator end")
	}()
	for {
		select {
		case <-s.endChan:
			return
		default:
			var email = []string{s.Randomizer.RandomString(10), "@", s.Randomizer.RandomString(10), ".ru"}
			element := TestTable{Password: s.Randomizer.RandomString(10), Email: strings.Join(email, ""), Active: false}
			s.Lock()
			s.StockWrite = append(s.StockWrite, element)
			s.Unlock()
			time.Sleep(time.Second * 1)
		}
	}

}
func (s *SqliteStock) workerDBS() {
	log.Println("workerDBS starting")
	//чтение из базы данных всех записей, что имеет статус Active:False и размещение их в StockWrite
	var records []TestTable
	if err := s.DB.Where(&TestTable{Active: false}).Find(&records).Error; err != nil {
		log.Fatal(err)
	}
	if len(records) > 0 {
		for _, x := range records {
			s.StockRead = append(s.StockRead, x)
		}
	}

	defer func() {
		s.Done()
		log.Println("workerDBS end")
	}()
	for {
		select {
		case <-s.endChan:
			return
		default:
			//проверка стока новых записей для проверки и внесения в базу данных + в сток для рабочих для обработки
			if len(s.StockWrite) > 0 {

				for i := 0; i < len(s.StockWrite); i++ {
					var element = s.StockWrite[0]
					s.StockWrite = append(s.StockWrite[:i], s.StockWrite[i+1:]...)
					i--
					fmt.Printf("Element: %v\n", element)

					switch element.Active {
					case false:
						var allrecords []TestTable
						if err := s.DB.Find(&allrecords).Error; err != nil {
							log.Fatal(err)
						}
						if len(allrecords) > 0 {
							for _, x := range allrecords {
								if x.Email == element.Email {
									fmt.Printf("DUBLICATE\n")
								} else {
									if err := s.DB.Create(&element).Error; err != nil {
										//log.Println(err)
									}
								}
							}
						}  else {
							if err := s.DB.Create(&element).Error; err != nil {
								//log.Println(err)
							}

						}

						s.Lock()
						s.StockRead = append(s.StockRead, element)
						s.Unlock()

					case true:

						if err := s.DB.Save(&element).Error; err != nil {
							//log.Println(err)
						}
					default:
						fmt.Printf("WRONG VALUE `ACTIVE`\n")
					}

				}
			}
			time.Sleep(time.Second * 1)

			////проверка на аттрибут
			////делаю выборку из базы данных всех записей
			//var allrecords []*TestTable
			//if err := s.DB.Find(&allrecords).Error; err != nil {
			//	log.Fatal(err)
			//}
			////обработка записей с базы данных и сверка с новой записью в стоке с новыми записями
			//for i:=0; i < len(s.StockWrite); i++ {
			//
			//	switch s.StockWrite[i].Active {
			//	case false:
			//		//элемент поступил от генератора, значит его надо проверить на дублирование в базе данных, можно через unique,
			//		// но тут ради эмуляции модели работы, если этого элемента нет в базе данных он заносится в базу данных и размещается
			//		// в стоке для рабочих для обработки
			//		// в дальнейшем при запуске программы надо считывать данные из базы данных всех элементов, что имеет статус false и заносить
			//		// их в рабочих сток
			//		if len(records) > 0 {
			//			for _, z := range allrecords {
			//				fmt.Printf(">>[%d]  %v\n",len(s.StockWrite), s.StockWrite[i])
			//				if z.Email == s.StockWrite[i].Email {
			//					log.Println("HAVE DUBLICATE")
			//				} else {
			//
			//					//заносим новую запись в базу данных
			//					log.Println("Write new Record")
			//					if err := s.DB.Create(&s.StockWrite[i]).Error; err != nil {
			//						log.Println(err)
			//					}
			//					s.Lock()
			//					//заносим элемент в рабочий сток
			//					s.StockRead = append(s.StockRead, s.StockWrite[i])
			//					//удаляю занесенный элемент
			//					s.StockWrite = append(s.StockWrite[:i], s.StockWrite[i+1:]...)
			//					s.Unlock()
			//					i--
			//				}
			//			}
			//
			//
			//
			//		} else {
			//			//заносим новую запись в базу данных
			//			log.Println("Write new Record[else]")
			//			if err := s.DB.Create(&s.StockWrite[i]).Error; err != nil {
			//				log.Println(err)
			//			}
			//			//заносим элемент в рабочий сток
			//			s.Lock()
			//			//заносим элемент в рабочий сток
			//			s.StockRead = append(s.StockRead, s.StockWrite[i])
			//			//удаляю занесенный элемент
			//			s.StockWrite = append(s.StockWrite[:i], s.StockWrite[i+1:]...)
			//			s.Unlock()
			//			i--
			//
			//		}
			//
			//	case true:
			//		//элемент от рабочего, кто обработал элемент из StockRead, просто элемент обновляется в базе данных
			//		if err := s.DB.Save(s.StockWrite[i]).Error; err != nil {
			//			log.Println(err)
			//		}
			//		s.Lock()
			//		s.StockWrite = append(s.StockWrite, s.StockWrite[i])
			//		s.StockWrite = append(s.StockWrite[:i], s.StockWrite[i+1:]...)
			//		s.Unlock()
			//		i--
			//
			//	default:
			//		log.Println("ERROR `ACTIVE` atribute")
			//	}
			//}

		}
	}
}
func (s *SqliteStock) workerExample(name string) {
	log.Println(fmt.Sprintf("worker `%s` starting", name))
	defer func() {
		s.Done()
		log.Println(fmt.Sprintf("worker `%s` end", name))
	}()
	for {
		select {
		case <-s.endChan:
			return
		default:
			//fmt.Printf("STOCKREAD: %v\n", s.StockRead)
			if len(s.StockRead) > 0 {
				s.Lock()
				var element = TestTable{}
				if len(s.StockRead) >= 1 {
					element = s.StockRead[0]
					s.StockRead = append(s.StockRead[:0], s.StockRead[1:]...)
				} else {
					element = s.StockRead[0]
				}
				s.Unlock()
				if element.Active == false {
					element.ExHash = "Some Example value" + s.Randomizer.RandomStringChoice(20, utils.Lhexdigits)
					element.Active = true
					s.Lock()
					s.StockWrite = append(s.StockWrite, element)
					s.Unlock()
					log.Println(fmt.Sprintf("worker `%s` add element %v\n", name, element))
				}


			} else {
				time.Sleep(time.Microsecond * 200)
			}

		}
	}

}

//---------------------------------------------------------------------------
//  run запуск проверки работы совместной работы в WAL режиме при 1 писателя и
//  множестве читателей
//---------------------------------------------------------------------------

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
		case <-chanEnd:
			return
		default:
			if trigger < countGenerator {
				trigger ++
				s.Lock()
				var email = strings.Join([]string{s.Randomizer.RandomString(10), "@", s.Randomizer.RandomString(10), ".ru"}, "")
				s.Stock = append(s.Stock, &TestTable{Email: email, Password: s.Randomizer.RandomString(10)})
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
		case <-chanEdn:
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
