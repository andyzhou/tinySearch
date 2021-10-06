package main

import (
	"fmt"
	"github.com/andyzhou/tinySearch"
	"github.com/andyzhou/tinySearch/define"
	"github.com/andyzhou/tinySearch/json"
	"log"
	"math/rand"
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
	//testClientSuggestDoc(client)

	//agg doc
	//testClientAggDoc(client)

	//query doc
	//testClientQueryDoc(client)

	//remove doc
	//testClientRemoveDoc(client)

	//sync doc
	testClientSyncDoc(client)
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
	//filter for age property
	filterAge := json.NewFilterField()
	filterAge.Field = "prop.age"
	filterAge.Kind = define.FilterKindNumericRange
	filterAge.MinFloatVal = float64(5)
	filterAge.MaxFloatVal = float64(100)

	//filter for city property
	filterCity := json.NewFilterField()
	filterCity.Kind = define.FilterKindMatch
	filterCity.Field = "prop.city"
	filterCity.Val = "beijing"

	optJson := json.NewQueryOptJson()
	optJson.HighLight = true
	optJson.AddFilter(filterAge, filterCity)
	resp, err := client.DocQuery(ServerIndexTag, optJson)
	if resp != nil {
		fmt.Println("resp size:", resp.Total)
	}
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
	var (
		docIdBegin, docIdEnd int64
	)
	docIdBegin = 1445641905684619264
	docIdEnd = 1445641905684619265
	for id := docIdBegin; id <= docIdEnd; id++ {
		addOneDoc(id, client)
	}
}

//add one doc
func addOneDoc(docId int64, client *tinySearch.Client)  {
	//init test doc json
	docIdStr := fmt.Sprintf("%v", docId)
	testDocJson := json.NewTestDocJson()
	testDocJson.Id = docId
	testDocJson.Title = fmt.Sprintf("工信处女干事每月经过下属科室都要亲口交代24口交换机等技术性器件的安装工作-%d", docId)
	testDocJson.Cat = "job"
	testDocJson.Price = 10.1
	testDocJson.Prop["age"] = docId
	testDocJson.Prop["city"] = "beijing"
	testDocJson.Num = int64(rand.Intn(100))
	testDocJson.Introduce = "The second one 你 中文测试中文 is even more interesting! 吃水果"
	testDocJson.CreateAt = time.Now().Unix()

	jsonByte, _ := testDocJson.Encode()

	//kv := make(map[string]interface{})
	//vv := json.NewBaseJson()
	//vv.DecodeSimple(jsonByte, kv)

	err := client.DocSync(ServerIndexTag, docIdStr, jsonByte)
	if err != nil {
		fmt.Printf("sync doc %d failed, err:%v\n", docId, err.Error())
	}else{
		fmt.Printf("sync doc %d succeed.\n", docId)
	}
}