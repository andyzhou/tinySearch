package testing

import (
	"fmt"
	"github.com/andyzhou/tinysearch"
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/example/json"
	tJson "github.com/andyzhou/tinysearch/json"
	"log"
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

//get one doc
func getOneDoc(docId int64, client *tinysearch.Client)  {
	docIds := []string{
		fmt.Sprintf("%v", docId),
	}
	_, err := client.DocGet(ServiceIndexTag, docIds...)
	if err != nil {
		log.Printf("get doc failed 1, err:%v\n", err.Error())
		return
	}
	//for _, jsonByte := range jsonByteSlice {
	//	hitDoc := tJson.NewHitDocJson()
	//	err = hitDoc.Decode(jsonByte)
	//	if err != nil {
	//		log.Printf("get doc failed 2, err:%v", err)
	//		continue
	//	}
	//}
}

//query doc
func queryDoc(client *tinysearch.Client) {
	//filter for city property
	//used for match multi term value, at least one match.
	filterProp := tJson.NewFilterField()
	filterProp.Field = "prop.city"
	filterProp.Kind = define.FilterKindTermsQuery
	filterProp.Terms = []string{
		"beijing", //matched
		"liaoning", //not matched
	}
	filterProp.IsMust = true

	//filter for tag
	filterTag := tJson.NewFilterField()
	filterTag.Kind = define.FilterKindMatch
	filterTag.Field = "tags"
	filterTag.Terms = []string{"job"}
	filterTag.IsMust = true

	//filter for cat
	filterCat := tJson.NewFilterField()
	filterCat.Kind = define.FilterKindTermsQuery
	filterCat.Field = "catPath"
	filterCat.Val = "1,2,0"
	filterCat.IsMust = true

	//filter for price
	filterPrice := tJson.NewFilterField()
	filterPrice.Kind = define.FilterKindNumericRange
	filterPrice.Field = "price"//"prop.city"
	filterPrice.MinFloatVal = 8.0
	filterPrice.MaxFloatVal = 12.0
	filterPrice.IsMust = true

	//filter for prefix
	filterPrefix := tJson.NewFilterField()
	filterPrefix.Kind = define.FilterKindPrefix
	filterPrefix.Field = "catPath"
	filterPrefix.Val = "1,2"
	filterPrefix.IsMust = true

	//filter for query
	filterQueryCat := tJson.NewFilterField()
	filterQueryCat.Kind = define.FilterKindMatch
	filterQueryCat.Field = "title"
	filterQueryCat.Val = "安装"
	filterQueryCat.IsMust = true

	optJson := tJson.NewQueryOptJson()
	//optJson.Fields = []string{"title", "cat"}
	//optJson.Key = "job-1"
	optJson.QueryKind = define.QueryKindOfMatchAll
	optJson.HighLight = true
	optJson.Filters = []*tJson.FilterField{
	}
	//optJson.NeedDocs = true
	optJson.Offset = 0
	optJson.Size = 1

	//doc query
	resp, err := client.DocQuery(ServiceIndexTag, optJson)
	if err != nil {
		log.Println("testClientQueryDoc failed, err:", err)
		return
	}
	if resp == nil {
		log.Println("testClientQueryDoc no any record")
		return
	}
	//log.Printf("hit docs:%v\n", resp.Total)

	//analyze result
	for _, jsonObj := range resp.Records {
		if jsonObj.OrgJson == nil {
			continue
		}
		//log.Printf("hitId:%v, score:%v, orgJson:%v\n", jsonObj.Id, jsonObj.Score, string(jsonObj.OrgJson))
		testJson := json.NewTestDocJson()
		err = testJson.Decode(jsonObj.OrgJson)
		if testJson == nil {
			continue
		}
		//log.Println("testClientQueryDoc rec:", testJson)
	}
}

//test get one doc
func TestGetDoc(t *testing.T) {
	docId := int64(1)
	getOneDoc(docId, client)
}

//test query doc
func TestQueryDoc(t *testing.T) {
	queryDoc(client)
}

//benchmark add doc
func BenchmarkAddDoc(b *testing.B) {
	start := int64(30000)
	for i := 0; i < b.N; i++ {
		docId := int64(i+1) + start
		addOneDoc(b, docId, client)
	}
}

//benchmark get doc
func BenchmarkGetDoc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		docId := int64(i+1)
		getOneDoc(docId, client)
	}
}

//benchmark query doc
func BenchmarkQueryDoc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		queryDoc(client)
	}
}