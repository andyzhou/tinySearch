package main

import (
	"fmt"
	"github.com/andyzhou/tinySearch/face"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
	service := face.NewService(RpcPort)

	//add node
	rpcAddr := fmt.Sprintf("%s:%d", RpcHost, RpcPort)
	service.AddNode(rpcAddr)

	//add index
	service.AddIndex(IndexPath, IndexTag)

	//start wait group
	wg.Add(1)
	fmt.Println("start example...")

	//testing
	go docTesting(service)

	wg.Wait()
	service.Quit()
	fmt.Println("stop example...")
}

//doc testing
func docTesting(service iface.IService)  {
	index := service.GetIndex(IndexTag)
	doc := service.GetDoc()
	if index == nil || doc == nil {
		return
	}

	//init test doc json
	docId := "1"
	testDocJson := json.NewTestDocJson()
	testDocJson.Id = docId
	testDocJson.Title = "test"
	testDocJson.Cat = "car"
	testDocJson.Price = 10.1
	testDocJson.Introduce = "this is test"
	testDocJson.CreateAt = time.Now().Unix()

	//add doc
	docJson := json.NewDocJson()
	docJson.Id = docId
	docJson.JsonObj = testDocJson

	//just add into local
	bRet := doc.AddDoc(index, docJson)
	fmt.Println("add doc result:", bRet)

	//add into batch nodes
	bRet = service.DocSync(IndexTag, docId, testDocJson.Encode())
	fmt.Println("sync doc result:", bRet)
}