package main

import (
	"github.com/andyzhou/tinySearch"
	"log"
	"sync"
)

/*
 * service example
 */

const (
	RpcPort = 6060
	IndexPath = "./search_data"
	IndexTag = "test"
)

func main() {
	var (
		wg sync.WaitGroup
	)

	//watch signal
	tinySearch.WatchSignal(&wg)

	//init service
	service := tinySearch.NewServiceWithPara(IndexPath, RpcPort)

	//add index
	service.AddIndex(IndexTag)

	//start wait group
	log.Printf("start server on port %v\n", RpcPort)
	wg.Add(1)
	wg.Wait()
	service.Quit()
	log.Printf("stop server on port %v\n", RpcPort)
}