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
	ServiceRpcPort = 6060
	ServiceIndexTag = "test"
	DocSuggesterTag = "test"
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
	node := fmt.Sprintf(":%d", ServiceRpcPort)
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
	//suggest key words
	//testClientSuggestKeywords(client)

	//agg doc
	//testClientAggDoc(client)

	//query doc
	//testClientQueryDoc(client)

	//remove doc
	//testClientRemoveDoc(client)

	//get doc
	testClientGetDoc(client)

	//add doc
	//testClientSyncDoc(client)

	//create index
	//testClientCreateIndex(client)
}

//test suggest key words
func testClientSuggestKeywords(client *tinysearch.Client) {
	//query kind support:
	//QueryKindOfMatchQuery
	//QueryKindOfPhrase
	//QueryKindOfPrefix
	//QueryKindOfMatchAll

	//if set key and query kind empty
	//will get top N suggest words
	optJson := json.NewQueryOptJson()
	//optJson.Key = "seco"
	//optJson.QueryKind = define.QueryKindOfPrefix //define.QueryKindOfPhrase //or QueryKindOfPrefix
	optJson.SuggestTag = DocSuggesterTag
	resp, err := client.DocSuggest(ServiceIndexTag, optJson)
	if err != nil {
		log.Println("testClientSuggestDoc failed, err:", err)
	}else{
		log.Println("testClientSuggestDoc resp:", resp)
		if resp != nil {
			for _, v := range resp.List {
				log.Printf("key:%v, count:%v\n", v.Key, v.Count)
			}
		}
	}
}

//test agg doc
func testClientAggDoc(client *tinysearch.Client) {
	optJson := json.NewQueryOptJson()
	optJson.Key = "工作"

	//one agg field on single field
	oneAggField := optJson.GenAggField()
	oneAggField.Field = "cat"
	oneAggField.Size = 10
	//add one agg field
	optJson.AddAggField(oneAggField)

	//one agg field on hashed field
	secondAggField := optJson.GenAggField()
	secondAggField.Field = "prop.city"
	secondAggField.Size = 10

	//add one agg field
	optJson.AddAggField(secondAggField)

	resp, err := client.DocAgg(ServiceIndexTag, optJson)
	if err != nil {
		log.Println("testClientAggDoc failed, err:", err)
	}else{
		log.Println("testClientAggDoc")
		for facetName, facetSlice := range resp.MapList {
			if facetSlice == nil {
				continue
			}
			for _, aggList := range facetSlice {
				log.Printf("facetName:%v, aggList:%v\n", facetName, aggList)
			}
		}
	}
}

//get doc
func testClientGetDoc(client *tinysearch.Client)  {
	docIds := []string{
		fmt.Sprintf("%v", 1),
		fmt.Sprintf("%v", 2),
	}
	jsonByteSlice, err := client.DocGet(ServiceIndexTag, docIds...)
	if err != nil {
		log.Println(err)
		return
	}

	for _, jsonByte := range jsonByteSlice {
		hitDoc := json.NewHitDocJson()
		err = hitDoc.Decode(jsonByte)
		if err != nil {
			continue
		}
		testDocJson := json.NewTestDocJson()
		err = testDocJson.Decode(hitDoc.OrgJson)
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
	filterProp := json.NewFilterField()
	filterProp.Field = "prop.city"
	filterProp.Kind = define.FilterKindTermsQuery
	filterProp.Terms = []string{
		"beijing", //matched
		"liaoning", //not matched
	}
	filterProp.IsMust = true

	//filter for tag
	filterTag := json.NewFilterField()
	filterTag.Kind = define.FilterKindMatch
	filterTag.Field = "tags"
	filterTag.Terms = []string{"job"}
	filterTag.IsMust = true

	//filter for poster
	filterPoster := json.NewFilterField()
	filterPoster.Kind = define.FilterKindMatch
	filterPoster.Field = "posterId"
	filterPoster.Val = fmt.Sprintf("%v", 2)
	filterPoster.IsMust = true

	////filter for price
	//filterPrice := json.NewFilterField()
	//filterPrice.Kind = define.FilterKindMatch
	//filterPrice.Field = "price"//"prop.city"
	//filterPrice.Val = "10.2"

	optJson := json.NewQueryOptJson()
	//optJson.SuggestTag = DocSuggesterTag
	//optJson.Key = "second"
	optJson.HighLight = true
	optJson.Filters = []*json.FilterField{
		filterPoster,
	}
	resp, err := client.DocQuery(ServiceIndexTag, optJson)
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
	err := client.DocRemove(ServiceIndexTag, docId)
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
	docIdBegin = 1
	docIdEnd = 2
	for id := docIdBegin; id <= docIdEnd; id++ {
		addOneDoc(id, client)
	}
}

//test create index
func testClientCreateIndex(client *tinysearch.Client) {
	err := client.CreateIndex(ServiceIndexTag)
	log.Printf("create index %v err:%v", ServiceIndexTag, err)
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
	testDocJson.Tags = []string{
		"car", "job",
	}
	testDocJson.PosterId = fmt.Sprintf("%v", docId)
	jsonByte, err := testDocJson.Encode()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = client.DocSync(ServiceIndexTag, docIdStr, jsonByte)
	if err != nil {
		log.Printf("sync doc %d failed, err:%v\n", docId, err.Error())
	}else{
		log.Printf("sync doc %d succeed.\n", docId)
	}
}
