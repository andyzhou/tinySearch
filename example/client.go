package main

import (
	genJson "encoding/json"
	"fmt"
	"github.com/andyzhou/tinysearch"
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/json"
	"log"
	"math/rand"
	"sync"
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

	//watch signal
	tinysearch.WatchSignal(&wg)

	//init client
	client := tinysearch.NewClient()

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
func testing(client *tinysearch.Client) {
	//suggest doc
	//testClientSuggestDoc(client)

	//agg doc
	//testClientAggDoc(client)

	//query doc
	testClientQueryDoc(client)

	//remove doc
	//testClientRemoveDoc(client)

	//get doc
	//testClientGetDoc(client)

	//add doc
	//testClientSyncDoc(client)

	//create index
	//testClientCreateIndex(client)
}

//test suggest doc
func testClientSuggestDoc(client *tinysearch.Client) {
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
func testClientAggDoc(client *tinysearch.Client) {
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
func testClientGetDoc(client *tinysearch.Client)  {
	docIds := []string{
		fmt.Sprintf("%v", 1),
		fmt.Sprintf("%v", 2),
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
func testClientQueryDoc(client *tinysearch.Client) {
	//filter for city property
	//used for match multi term value, at least one match.
	filterAge := json.NewFilterField()
	filterAge.Field = "prop.city"
	filterAge.Kind = define.FilterKindTermsQuery
	filterAge.Terms = []string{
		"beijing", //matched
		"liaoning", //not matched
	}
	filterAge.IsMust = true
	////filter for city property
	//filterCity := json.NewFilterField()
	//filterCity.Kind = define.FilterKindMatch
	//filterCity.Field = "cat"//"prop.city"
	//filterCity.Val = "job"
	//
	////filter for price
	//filterPrice := json.NewFilterField()
	//filterPrice.Kind = define.FilterKindMatch
	//filterPrice.Field = "price"//"prop.city"
	//filterPrice.Val = "10.2"

	optJson := json.NewQueryOptJson()
	//optJson.Key = "chinese"
	optJson.HighLight = true
	optJson.Filters = []*json.FilterField{
		filterAge,
	}
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
		log.Println("jsonObj:", string(jsonObj.OrgJson))
		//testJson := json.NewTestDocJson()
		//err = testJson.Decode(jsonObj.OrgJson)
		//if err != nil {
		//	//fmt.Println(string(jsonObj.OrgJson))
		//	//fmt.Println(err.Error())
		//	continue
		//}
		//log.Println("testClientQueryDoc rec:", testJson)
	}
}

//test remove doc
func testClientRemoveDoc(client *tinysearch.Client) {
	docId := "4"
	err := client.DocRemove(ServerIndexTag, docId)
	if err != nil {
		log.Println("remove doc failed, err:", err.Error())
	}else{
		log.Println("remove doc succeed.")
	}
}

//test sync doc
func testClientSyncDoc(client *tinysearch.Client) {
	var (
		docIdBegin, docIdEnd int64
	)
	docIdBegin = 2
	docIdEnd = 2
	for id := docIdBegin; id <= docIdEnd; id++ {
		addOneDoc(id, client)
	}
}

//test create index
func testClientCreateIndex(client *tinysearch.Client) {
	err := client.CreateIndex(ServerIndexTag)
	log.Printf("create index %v err:%v", ServerIndexTag, err)
}

//add one doc
func addOneDoc(docId int64, client *tinysearch.Client) {
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
