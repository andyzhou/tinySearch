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
 * face for example client
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

const (
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
	service := tinySearch.NewSearch(RpcPort)

	//add node
	rpcAddr := fmt.Sprintf(":%d", RpcPort)
	service.AddNode(rpcAddr)

	//add index
	service.AddIndex(IndexPath, IndexTag)

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
func docTesting(
		service iface.ISearch,
	)  {
	var (
		bRet bool
	)

	//get relate face
	index := service.GetIndex(IndexTag)
	doc := service.GetDoc()
	query := service.GetQuery()
	agg := service.GetAgg()
	if index == nil || doc == nil {
		return
	}

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

	//add doc into local
	//err := doc.AddDoc(index, docId, testDocJson)
	//fmt.Println("add doc result:", err)

	//remove doc from local
	//bRet = doc.RemoveDoc(index, docId)
	//fmt.Println("remove doc result:", bRet)

	//remove doc from batch nodes
	//bRet = service.DocRemove(IndexTag, docId)
	//fmt.Println("remove doc result:", bRet)

	//query opt
	queryOptJson := json.NewQueryOptJson()

	//term query
	//queryOptJson.QueryKind = define.QueryKindOfTerm
	//queryOptJson.TermPara = json.TermQueryPara{
	//	Field:"cat",
	//	Val:"car",
	//}

	//key query
	//queryOptJson.Key = "car"
	//queryOptJson.Fields = []string{
	//	"cat",
	//}

	//setup filter
	filterOne := &json.FilterField{
		Kind:define.FilterKindMatch,
		Field:"cat",
		Val:"car",
	}

	filterTwo := &json.FilterField{
		Kind:define.FilterKindMatch,
		Field:"title",
		Val:"test1",
	}
	queryOptJson.AddFilter(filterOne, filterTwo)

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

	//agg doc
	queryOptJson.AggField = &json.AggField{
		Field:"cat",
		Size:10,
	}
	aggResult, _ := agg.GetAggList(index, queryOptJson)
	fmt.Println("aggResult:", aggResult)
}