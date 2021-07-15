package main

import (
	"fmt"
	"github.com/andyzhou/tinySearch"
	"github.com/andyzhou/tinySearch/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

/*
 * client example
 */

const (
	ServerRpcPort = 6060
	ServerIndexTag = "test"
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

	//init client
	client := tinySearch.NewClient()

	//add node
	node := fmt.Sprintf(":%d", ServerRpcPort)
	client.AddNodes(node)

	wg.Add(1)
	fmt.Println("client start..")

	//testing
	testClientSyncDoc(client)

	wg.Wait()
	fmt.Println("client stop..")
}

//test sync doc
func testClientSyncDoc(client *tinySearch.Client) {
	//init test doc json
	docId := "4"
	testDocJson := json.NewTestDocJson()
	testDocJson.Id = docId
	testDocJson.Title = "test-4"
	testDocJson.Cat = "car"
	testDocJson.Price = 10.1
	testDocJson.Num = 20
	testDocJson.Introduce = "this is test-1"
	testDocJson.CreateAt = time.Now().Unix()

	err := client.DocSync(ServerIndexTag, docId, testDocJson.Encode())
	if err != nil {
		fmt.Println("sync doc failed, err:", err.Error())
	}else{
		fmt.Println("sync doc succeed.")
	}
}