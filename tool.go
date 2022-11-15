package tinysearch

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/*
 * inter tool func
 */

//watch signal
func WatchSignal(wg *sync.WaitGroup) bool {
	//check
	if wg == nil {
		return false
	}

	//try catch signal
	c := make(chan os.Signal, 1)
	signal.Notify(
		c,
		os.Kill,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGABRT,
	)

	//signal snatch
	go func(wg *sync.WaitGroup) {
		var (
			needQuit bool
		)
		for {
			if needQuit {
				break
			}
			select {
			case s := <- c:
				log.Printf("Get signal of %v", s.String())
				wg.Done()
				needQuit = true
			}
		}
	}(wg)
	return true
}
