package goroutines

import (
	"sync"
	"log"
	"fmt"
	"time"
)

type (
	Core struct {
		sync.RWMutex
		sync.WaitGroup
		theEnd chan bool
	}
)
func NewCore() *Core {
	n := &Core{
		theEnd: make(chan bool),
	}
	return n
}
func (c *Core) Run(count int) {
	for x:=0; x < count; x++ {
		c.Add(1)
		go c.worker(fmt.Sprintf("W#%d", x))
	}
	var  timerEnd  = time.NewTimer(time.Second * 10 )
	for {
		select {

		}
	}

}
func (c *Core) worker(name string) {
	log.Println(fmt.Sprintf("worker %v start", name))
	defer func(){
		log.Println(fmt.Sprintf("worker %v end", name))
		c.Done()
	}()
	for {
		select {
		case <- c.theEnd:
			return
		default:
			log.Println(fmt.Sprintf("worker %v make some work and going to sleep", name))
			time.Sleep(time.Second * 1)
		}
	}
}