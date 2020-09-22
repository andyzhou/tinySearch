package main

import (
	"fmt"
	"github.com/andyzhou/tinySearch/face"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/*
 * face for example client
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

const (
	RpcHost = "127.0.0.1"
	RpcPort = 6060
	IndexPath = "/data/test"
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
	service := face.NewService(RpcPort)

	//add node
	rpcAddr := fmt.Sprintf("%s:%d", RpcHost, RpcPort)
	service.AddNode(rpcAddr)

	//add index
	service.AddIndex(IndexPath, "test")

	//start wait group
	wg.Add(1)
	fmt.Println("start example...")

	wg.Wait()
	fmt.Println("stop example...")
}