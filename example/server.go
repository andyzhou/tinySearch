package main

import (
	"fmt"
	"github.com/andyzhou/tinySearch"
	"github.com/andyzhou/tinySearch/define"
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
 * face for example service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

const (
	RpcPort = 6060
	IndexPath = "/data/search"
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
	service := tinySearch.NewSearch(IndexPath, RpcPort)

	//add node
	rpcAddr := fmt.Sprintf(":%d", RpcPort)
	service.AddNode(rpcAddr)

	//add index
	service.AddIndex(IndexTag)

	//start wait group
	fmt.Println("start example...")

	wg.Add(1)

	//testing
	go docTesting(service)

	wg.Wait()
	service.Quit()
	fmt.Println("stop example...")
}

//doc testing
func docTesting(service iface.ISearch)  {
	if service == nil {
		return
	}

	//test sync doc
	//testSyncDoc(service)

	//test query doc
	//testQuery(service)

	//test agg
	//testAgg(service)

	//test suggest
	testSuggest(service)
}

//test suggest
func testSuggest(service iface.ISearch) {
	//get relate face
	suggest := service.GetSuggest()

	//opt
	optJson := json.NewSuggestOptJson()
	optJson.Key = "test"

	//query
	rec := suggest.GetSuggest(optJson)
	fmt.Println("rec:", rec.Total)
}

//test agg
func testAgg(service iface.ISearch) {
	//get relate face
	index := service.GetIndex(IndexTag)
	agg := service.GetAgg()

	//query opt
	queryOptJson := json.NewQueryOptJson()

	//key query
	queryOptJson.Key = "test"
	//queryOptJson.Fields = []string{
	//	"cat",
	//}

	//setup filter
	filterOne := &json.FilterField{
		Kind:define.FilterKindMatch,
		Field:"cat",
		Val:"car",
	}

	//filterTwo := &json.FilterField{
	//	Kind:define.FilterKindMatch,
	//	Field:"title",
	//	Val:"test1",
	//}
	queryOptJson.AddFilter(filterOne)

	//agg doc
	queryOptJson.AggField = &json.AggField{
		Field:"cat",
		Size:10,
	}
	aggResult, _ := agg.GetAggList(index, queryOptJson)
	fmt.Println("aggResult:", aggResult)
}

//test query
func testQuery(service iface.ISearch) {
	var (
		bRet bool
	)

	//get relate face
	index := service.GetIndex(IndexTag)

	//get query face
	query := service.GetQuery()

	//query opt
	queryOptJson := json.NewQueryOptJson()

	//term query
	//queryOptJson.QueryKind = define.QueryKindOfTerm
	//queryOptJson.TermPara = json.TermQueryPara{
	//	Field:"cat",
	//	Val:"car",
	//}

	//key query
	queryOptJson.Key = "test"
	//queryOptJson.Fields = []string{
	//	"cat",
	//}

	//setup filter
	filterOne := &json.FilterField{
		Kind:define.FilterKindMatch,
		Field:"cat",
		Val:"car",
	}

	//filterTwo := &json.FilterField{
	//	Kind:define.FilterKindMatch,
	//	Field:"title",
	//	Val:"test1",
	//}
	queryOptJson.AddFilter(filterOne)

	//query batch doc
	result, _ := query.Query(index, queryOptJson)
	fmt.Println("result:", result)
	if result != nil {
		for _, hitObj := range result.Records {
			testJson := json.NewTestDocJson()
			jsonStr := string(hitObj.OrgJson)
			bRet = testJson.Decode(hitObj.OrgJson)
			if !bRet {
				continue
			}
			fmt.Println("jsonStr:", jsonStr)
			//fmt.Println("testJson:", testJson)
		}
	}
}

//test sync doc
func testSyncDoc(service iface.ISearch) {
	//init test doc json
	docId := "2"
	testDocJson := json.NewTestDocJson()
	testDocJson.Id = docId
	testDocJson.Title = "test-1"
	testDocJson.Cat = "car"
	testDocJson.Price = 10.1
	testDocJson.Num = 20
	testDocJson.Introduce = "this is test-1"
	testDocJson.CreateAt = time.Now().Unix()

	//sync doc
	err := service.DocSync(IndexTag, docId, testDocJson.Encode())
	fmt.Println("doc sync, err:", err)
}