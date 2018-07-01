package goroutines

import "sync"
import (
	"github.com/spouk/gocheats/utils"
	"fmt"
	"time"
	"log"
)

type Manager struct {
	sync.RWMutex
	sync.WaitGroup
	Stock   []string
	counter int64
	r       *utils.Randomizer
	endChan chan bool
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
		case <- m.endChan:
			return
		default:
			m.Lock()
			m.Stock = append(m.Stock, m.r.RandomStringChoice(30, utils.Lhexdigits))
			m.Unlock()
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
				fmt.Printf(fmt.Sprintf("[%d] %v `%s`\n", m.counter, name, element))
				time.Sleep(time.Second * 1)
			} else {
				log.Println(fmt.Sprintf("%v STOCK EMPTY", name))
				time.Sleep(time.Second * 2)
			}
		}
	}
}
