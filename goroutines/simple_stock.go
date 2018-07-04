package goroutines

import (
	"github.com/spouk/gocheats/utils"
	"fmt"
	"time"
	"log"
	"sync"
)

type Manager struct {
	sync.RWMutex
	sync.WaitGroup
	Stock    []Static
	EndStock []Static
	counter  int64
	r        *utils.Randomizer
	endChan  chan bool
}
type Static struct {
	Name   string
	Hash   string
	Active bool
}

func NewManager() *Manager {
	m := &Manager{
		r:       utils.NewRandomize(),
		endChan: make(chan bool),
	}
	return m
}
func (m *Manager) Run(count, stockSize int, timeTrigger int) {
	//for x := 0; x <= stockSize; x++ {
	//	m.Stock = append(m.Stock, m.r.RandomString(20))
	//}
	m.Add(1)
	go m.generator()

	for x := 0; x <= count; x++ {
		m.Add(1)
		go m.worker(fmt.Sprintf("worker#%d", x))
	}
	m.counter = 0
	<-time.NewTimer(time.Second * time.Duration(timeTrigger)).C
	close(m.endChan)
	m.Wait()
	fmt.Printf("END")
}
func (m *Manager) generator() {
	defer m.Done()
	for {
		select {
		case <-m.endChan:
			return
		default:
			//make new element to .Stock
			m.Lock()
			m.Stock = append(m.Stock, Static{Name: m.r.RandomStringChoice(30, utils.Lhexdigits)})
			m.Unlock()
			//check new elements in .EndStock
			if len(m.EndStock) > 0 {
				var element = m.EndStock[0]
				m.Lock()
				m.EndStock = append(m.EndStock[:0], m.EndStock[1:]...)
				m.Unlock()
				fmt.Printf("[generator][out] %v\n", element)
			}
			time.Sleep(time.Microsecond * 5000)
		}
	}
}
func (m *Manager) worker(name string) {
	defer m.Done()
	for {
		select {
		case <-m.endChan:
			return
		default:
			if len(m.Stock) > 0 {
				m.Lock()
				element := m.Stock[0]
				m.Stock = append(m.Stock[:0], m.Stock[1:]...)
				m.counter++
				m.Unlock()

				element.Active = true
				element.Hash = m.r.RandomString(20)
				m.Lock()
				m.EndStock = append(m.EndStock, element)
				m.Unlock()

				fmt.Printf(fmt.Sprintf("[%d] %v `%s`\n", m.counter, name, element))
				time.Sleep(time.Second * 1)
			} else {
				log.Println(fmt.Sprintf("%v STOCK EMPTY", name))
				time.Sleep(time.Second * 2)
			}
		}
	}
}
