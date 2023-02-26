package main

import (
	"github.com/andyzhou/tinysearch"
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
	tinysearch.WatchSignal(&wg)

	//init service
	service := tinysearch.NewServiceWithPara(IndexPath, RpcPort)

	//set relate path
	service.SetDataPath(IndexPath)
	service.SetDictFile("")

	//add index
	service.AddIndex(IndexTag)

	//start wait group
	log.Printf("start server on port %v\n", RpcPort)
	wg.Add(1)
	wg.Wait()
	service.Quit()
	log.Printf("stop server on port %v\n", RpcPort)
}