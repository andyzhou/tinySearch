package main

import (
	"github.com/andyzhou/tinysearch"
	"log"
	"os"
	"sync"
)

/*
 * service example
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

const (
	RpcPort = 6160
	IndexPath = "./search_data"
	IndexTag = "test"
	SuggesterTag = "test"
)

//hook for add doc
func hookForAddDoc(jsonBytes []byte) error {
	log.Println("jsonBytes:", jsonBytes)
	return nil
}

func main() {
	var (
		wg sync.WaitGroup
	)

	//watch signal
	tinysearch.WatchSignal(&wg)

	//format service para
	servicePara := &tinysearch.ServicePara{
		DataPath: IndexPath,
		RpcPort: RpcPort,
		AddDocQueueMode: true,
	}

	//init service
	service := tinysearch.NewServiceWithPara(servicePara)

	//set relate path
	service.SetDataPath(IndexPath)
	//service.SetDictFile("")

	//set hook for add doc
	service.SetHookForAddDoc(hookForAddDoc)

	//add index
	err := service.AddIndex(IndexTag)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	//register suggester tag
	service.GetSuggest().RegisterSuggest(SuggesterTag)

	//start wait group
	log.Printf("start server on port %v\n", RpcPort)
	wg.Add(1)
	wg.Wait()
	service.Quit()
	log.Printf("stop server on port %v\n", RpcPort)
}