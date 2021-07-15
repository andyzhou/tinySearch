package main

import (
	"fmt"
	"github.com/andyzhou/tinySearch"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/*
 * service example
 */

const (
	RpcPort = 6060
	IndexPath = "/data/search"
	IndexTag = "test"
)

func main() {
	var (
		wg sync.WaitGroup
	)

	//try catch signal
	c := make(chan os.Signal, 1)
	signal.Notify(
		c,
		os.Kill,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGABRT,
	)

	///signal snatch
	go func(wg *sync.WaitGroup) {
		var needQuit bool
		for {
			if needQuit {
				break
			}
			select {
			case s := <- c:
				log.Println("Get signal of ", s.String())
				wg.Done()
				needQuit = true
			}
		}
	}(&wg)

	//init service
	service := tinySearch.NewService(IndexPath, RpcPort)

	//add index
	service.AddIndex(IndexTag)

	//start wait group
	fmt.Printf("start server on port %v\n", RpcPort)

	wg.Add(1)

	wg.Wait()
	service.Quit()
	fmt.Println("stop server...")
}
