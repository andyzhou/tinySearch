package testing

import (
	"fmt"
	"github.com/andyzhou/tinysearch"
	"github.com/andyzhou/tinysearch/example/json"
	tJson "github.com/andyzhou/tinysearch/json"
	"math/rand"
	"testing"
	"time"
)

const (
	ServiceRpcPort = 6160
	ServiceIndexTag = "test"
)

var (
	client *tinysearch.Client
)

//init
func init() {
	//init client
	client = tinysearch.NewClient()

	//add node
	node := fmt.Sprintf(":%d", ServiceRpcPort)
	client.AddNodes(node)
}

//add new doc
func addOneDoc(b *testing.B, docId int64, client *tinysearch.Client) {
	//init test doc json
	docIdStr := fmt.Sprintf("%v", docId)
	testDocJson := json.NewTestDocJson()
	testDocJson.Id = docId
	testDocJson.Title = fmt.Sprintf("工信处女干事每月件的安装工作-%d", docId)
	testDocJson.Cat = "job-1"
	testDocJson.CatPath ="1,2,0"
	testDocJson.Price = 10.2
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
		b.Logf("sync doc %d failed, err:%v\n", docId, err.Error())
	}
}

//get doc
func getDoc(b *testing.B, docId int64, client *tinysearch.Client)  {
	docIds := []string{
		fmt.Sprintf("%v", docId),
	}
	jsonByteSlice, err := client.DocGet(ServiceIndexTag, docIds...)
	if err != nil {
		b.Logf("get doc failed 1, err:%v\n", err.Error())
		return
	}

	for _, jsonByte := range jsonByteSlice {
		hitDoc := tJson.NewHitDocJson()
		err = hitDoc.Decode(jsonByte)
		if err != nil {
			b.Logf("get doc failed 2, err:%v", err)
			continue
		}
	}
}

//benchmark add doc
func BenchmarkAddDoc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		docId := int64(i+1)
		addOneDoc(b, docId, client)
	}
}

//benchmark get doc
func BenchmarkGetDoc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		docId := int64(i+1)
		getDoc(b, docId, client)
	}
}