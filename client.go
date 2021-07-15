package tinySearch

import (
	"errors"
	"github.com/andyzhou/tinySearch/face"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	"log"
	"sync"
)

/*
 * client api
 */

//query opt kind
const (
	QueryOptKindOfGen = iota
	QueryOptKindOfAgg
	QueryOptKindOfSuggest
)

//face info
type Client struct {
	rpcClients map[string]iface.IRpcClient
	sync.RWMutex
}

//construct
func NewClient() *Client {
	//self init
	self := &Client{
		rpcClients:make(map[string]iface.IRpcClient),
	}
	return self
}

//quit
func (f *Client) Quit() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client:Quit panic, err:", err)
		}
	}()
	if f.rpcClients != nil {
		for _, client := range f.rpcClients {
			client.Quit()
		}
	}
}

//suggest doc
func (f *Client) DocSuggest(
					indexTag string,
					optJson *json.QueryOptJson,
				) (*json.SuggestsJson, error) {
	var (
		bRet bool
	)

	//check
	if indexTag == "" || optJson == nil {
		return nil, errors.New("invalid parameter")
	}

	//get rpc client
	client := f.getClient()
	if client == nil {
		return nil, errors.New("can't get active rpc client")
	}

	//call api
	jsonByteSlice, _, err := client.DocQuery(
		QueryOptKindOfSuggest,
		indexTag,
		optJson.Encode(),
	)
	if err != nil {
		return nil, err
	}

	//analyze result
	if jsonByteSlice == nil || len(jsonByteSlice) <= 0 {
		return nil, nil
	}

	//format result
	suggestsJson := json.NewSuggestsJson()
	for _, jsonByte := range jsonByteSlice {
		suggestJson := json.NewSuggestJson()
		bRet = suggestJson.Decode(jsonByte)
		if !bRet {
			continue
		}
		suggestsJson.AddObj(suggestJson)
	}
	return suggestsJson, nil
}

//agg doc
func (f *Client) DocAgg(
					indexTag string,
					optJson *json.QueryOptJson,
				) (*json.AggregatesJson, error) {
	var (
		bRet bool
	)

	//check
	if indexTag == "" || optJson == nil {
		return nil, errors.New("invalid parameter")
	}

	//get rpc client
	client := f.getClient()
	if client == nil {
		return nil, errors.New("can't get active rpc client")
	}

	//call api
	jsonByteSlice, _, err := client.DocQuery(
									QueryOptKindOfAgg,
									indexTag,
									optJson.Encode(),
								)
	if err != nil {
		return nil, err
	}

	//analyze result
	if jsonByteSlice == nil || len(jsonByteSlice) <= 0 {
		return nil, nil
	}

	//format result
	aggJsonSlice := json.NewAggregatesJson()
	for _, jsonByte := range jsonByteSlice {
		aggJson := json.NewAggregateJson()
		bRet = aggJson.Decode(jsonByte)
		if !bRet {
			continue
		}
		aggJsonSlice.AddObj(aggJson)
	}
	return aggJsonSlice, nil
}

//query doc
func (f *Client) DocQuery(
					indexTag string,
					optJson *json.QueryOptJson,
				) (*json.SearchResultJson, error) {
	var (
		bRet bool
	)

	//check
	if indexTag == "" || optJson == nil {
		return nil, errors.New("invalid parameter")
	}

	//get rpc client
	client := f.getClient()
	if client == nil {
		return nil, errors.New("can't get active rpc client")
	}

	//call api
	jsonByteSlice, hitDocs, err := client.DocQuery(
								QueryOptKindOfGen,
								indexTag,
								optJson.Encode(),
							)
	if err != nil {
		return nil, err
	}

	//analyze result
	if jsonByteSlice == nil || len(jsonByteSlice) <= 0 {
		return nil, nil
	}

	//format result
	queryResultJsons := json.NewSearchResultJson()
	queryResultJsons.Total = uint64(hitDocs)
	for _, jsonByte := range jsonByteSlice {
		hitDocJson := json.NewHitDocJson()
		bRet = hitDocJson.Decode(jsonByte)
		if !bRet {
			continue
		}
		queryResultJsons.AddDoc(hitDocJson)
	}
	return queryResultJsons, nil
}

//remove doc
func (f *Client) DocRemove(
					indexTag string,
					docIds ...string,
				) error {
	//check
	if indexTag == "" || docIds == nil {
		return errors.New("invalid parameter")
	}
	//get rpc client
	client := f.getClient()
	if client == nil {
		return errors.New("can't get active rpc client")
	}
	bRet := client.DocRemove(indexTag, docIds)
	if !bRet {
		return errors.New("doc remove failed")
	}
	return nil
}

//add sync
//used for add, sync doc
func (f *Client) DocSync(
					indexTag, docId string,
					docJson []byte,
				) error {
	//check
	if indexTag == "" || docId == "" || docJson == nil {
		return errors.New("invalid parameter")
	}
	//get rpc client
	client := f.getClient()
	if client == nil {
		return errors.New("can't get active rpc client")
	}
	bRet := client.DocSync(indexTag, docId, docJson)
	if !bRet {
		return errors.New("doc sync failed")
	}
	return nil
}

//add search service nodes
func (f *Client) AddNodes(nodes ... string) bool {
	//check
	if nodes == nil || len(nodes) <= 0 {
		return false
	}
	//check and init new rpc client
	for _, node := range nodes {
		//check
		_, ok := f.rpcClients[node]
		if ok {
			continue
		}
		//create new rpc client
		rpcClient := face.NewRpcClient(node)
		f.rpcClients[node] = rpcClient
	}
	return true
}

//////////////
//private func
//////////////

//get rand active rpc client
func (f *Client) getClient() iface.IRpcClient {
	if f.rpcClients == nil {
		return nil
	}
	for _, client := range f.rpcClients {
		if client.IsActive() {
			return client
		}
	}
	return nil
}