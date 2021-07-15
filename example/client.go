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
	testing(client)

	wg.Wait()
	fmt.Println("client stop..")
}

//testing
func testing(client *tinySearch.Client) {
	//suggest doc
	testClientSuggestDoc(client)

	//agg doc
	//testClientAggDoc(client)

	//query doc
	//testClientQueryDoc(client)

	//remove doc
	//testClientRemoveDoc(client)

	//sync doc
	//testClientSyncDoc(client)
}

//test suggest doc
func testClientSuggestDoc(client *tinySearch.Client) {
	optJson := json.NewQueryOptJson()
	optJson.Key = "te"
	resp, err := client.DocSuggest(ServerIndexTag, optJson)
	fmt.Println("resp:", resp)
	fmt.Println("err:", err)
}

//test agg doc
func testClientAggDoc(client *tinySearch.Client) {
	optJson := json.NewQueryOptJson()
	optJson.Key = "test"
	optJson.AggField = &json.AggField{
		Field:"cat",
		Size:10,
	}
	resp, err := client.DocAgg(ServerIndexTag, optJson)
	fmt.Println("resp:", resp)
	fmt.Println("err:", err)
}

//test query doc
func testClientQueryDoc(client *tinySearch.Client) {
	optJson := json.NewQueryOptJson()
	optJson.Key = "test"
	resp, err := client.DocQuery(ServerIndexTag, optJson)
	fmt.Println("resp:", resp)
	fmt.Println("err:", err)
}

//test remove doc
func testClientRemoveDoc(client *tinySearch.Client) {
	docId := "4"
	err := client.DocRemove(ServerIndexTag, docId)
	if err != nil {
		fmt.Println("remove doc failed, err:", err.Error())
	}else{
		fmt.Println("remove doc succeed.")
	}
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