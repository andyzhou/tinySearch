package main

import (
	genJson "encoding/json"
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
	testClientAggDoc(client)

	//query doc
	testClientQueryDoc(client)

	//remove doc
	//testClientRemoveDoc(client)

	//get doc
	//testClientGetDoc(client)

	//testClientSyncDoc(client)
}

//test suggest doc
func testClientSuggestDoc(client *tinySearch.Client) {
	optJson := json.NewQueryOptJson()
	optJson.Key = "te"
	resp, err := client.DocSuggest(ServerIndexTag, optJson)
	if err != nil {
		log.Println("testClientSuggestDoc failed, err:", err)
	}else{
		log.Println("testClientSuggestDoc resp:", resp)
	}
}

//test agg doc
func testClientAggDoc(client *tinySearch.Client) {
	optJson := json.NewQueryOptJson()
	optJson.Key = "工作"
	optJson.AggField = &json.AggField{
		Field:"cat",
		Size:10,
	}
	resp, err := client.DocAgg(ServerIndexTag, optJson)
	if err != nil {
		log.Println("testClientAggDoc failed, err:", err)
	}else{
		log.Println("testClientAggDoc")
		for _, v := range resp.List {
			log.Printf("name:%v, count:%v\n", v.Name, v.Count)
		}
	}
}


//get doc
func testClientGetDoc(client *tinySearch.Client)  {
	docIds := []string{
		fmt.Sprintf("%v", 1445641905684619264),
	}
	jsonByteSlice, err := client.DocGet(ServerIndexTag, docIds...)
	if err != nil {
		log.Println(err)
		return
	}

	for _, jsonByte := range jsonByteSlice {
		testDocJson := json.NewTestDocJson()
		err = testDocJson.Decode(jsonByte)
		if err != nil {
			log.Printf("testClientGetDoc failed, err:%v", err)
		}else{
			log.Printf("testClientGetDoc result:%v", testDocJson)
		}
	}
}

//test query doc
func testClientQueryDoc(client *tinySearch.Client) {
	//filter for age property
	//filterAge := json.NewFilterField()
	//filterAge.Field = "prop.age"
	//filterAge.Kind = define.FilterKindNumericRange
	//filterAge.MinFloatVal = float64(5)
	//filterAge.MaxFloatVal = float64(100)

	//filter for city property
	filterCity := json.NewFilterField()
	filterCity.Kind = define.FilterKindMatch
	filterCity.Field = "cat"//"prop.city"
	filterCity.Val = "job"

	//filter for price
	filterPrice := json.NewFilterField()
	filterPrice.Kind = define.FilterKindMatch
	filterPrice.Field = "price"//"prop.city"
	filterPrice.Val = "10.2"

	optJson := json.NewQueryOptJson()
	optJson.Key = "安装"
	optJson.HighLight = true
	//optJson.AddFilter(filterPrice)
	resp, err := client.DocQuery(ServerIndexTag, optJson)
	if err != nil {
		log.Println("testClientQueryDoc failed, err:", err)
		return
	}
	if resp == nil {
		log.Println("testClientQueryDoc no any record")
		return
	}

	//analyze result
	for _, jsonObj := range resp.Records {
		testJson := json.NewTestDocJson()
		err = testJson.Decode(jsonObj.OrgJson)
		if err != nil {
			//fmt.Println(string(jsonObj.OrgJson))
			//fmt.Println(err.Error())
			continue
		}
		log.Println("testClientQueryDoc rec:", testJson)
	}
}

//test remove doc
func testClientRemoveDoc(client *tinySearch.Client) {
	docId := "4"
	err := client.DocRemove(ServerIndexTag, docId)
	if err != nil {
		log.Println("remove doc failed, err:", err.Error())
	}else{
		log.Println("remove doc succeed.")
	}
}

//test sync doc
func testClientSyncDoc(client *tinySearch.Client) {
	var (
		docIdBegin, docIdEnd int64
	)
	docIdBegin = 1
	docIdEnd = 2
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
	testDocJson.Title = fmt.Sprintf("工信处女干事每月件的安装工作-%d", docId)
	testDocJson.Cat = "job"
	testDocJson.Price = genJson.Number(fmt.Sprintf("%v", 10.2))
	testDocJson.Prop["age"] = docId
	testDocJson.Prop["city"] = "beijing"
	testDocJson.Num = int64(rand.Intn(100))
	testDocJson.Introduce = "The second one 你 中文re interesting! 吃水果"
	testDocJson.CreateAt = time.Now().Unix()

	tagA := "car"
	tagB := "job"

	testDocJson.Tags[tagA] = 1
	testDocJson.Tags[tagB] = 1

	jsonByte, err := testDocJson.Encode()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = client.DocSync(ServerIndexTag, docIdStr, jsonByte)
	if err != nil {
		log.Printf("sync doc %d failed, err:%v\n", docId, err.Error())
	}else{
		log.Printf("sync doc %d succeed.\n", docId)
	}
}