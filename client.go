package tinySearch

import (
	"errors"
	"fmt"
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
	resultJson := json.NewSuggestsJson()
	bRet = resultJson.Decode(jsonByteSlice[0])
	if !bRet {
		return nil, errors.New("invalid json byte data")
	}
	return resultJson, nil
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
	resultJson := json.NewAggregatesJson()
	bRet = resultJson.Decode(jsonByteSlice[0])
	if !bRet {
		return nil, errors.New("invalid json byte data")
	}
	return resultJson, nil
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
	jsonByteSlice, _, err := client.DocQuery(
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
	resultJson := json.NewSearchResultJson()
	bRet = resultJson.Decode(jsonByteSlice[0])
	if !bRet {
		return nil, errors.New("invalid json byte data")
	}
	return resultJson, nil
}

//remove doc
func (f *Client) DocRemove(
					indexTag string,
					docIds ...string,
				) error {
	var (
		bRet bool
	)
	//check
	if indexTag == "" || docIds == nil {
		return errors.New("invalid parameter")
	}
	if f.rpcClients == nil {
		return errors.New("no any active rpc client")
	}
	//run on all rpc clients
	succeed := 0
	failed := 0
	for _, client := range f.rpcClients {
		if !client.IsActive() {
			failed++
			continue
		}
		bRet = client.DocRemove(indexTag, docIds...)
		if bRet {
			succeed++
		}else{
			failed++
		}
	}
	if failed > 0 {
		return errors.New(fmt.Sprintf("failed:%v, succeed:%v", failed, succeed))
	}
	return nil
}

//add sync
//used for add, sync doc, run on all nodes
func (f *Client) DocSync(
					indexTag, docId string,
					docJson []byte,
				) error {
	var (
		bRet bool
	)
	//check
	if indexTag == "" || docId == "" || docJson == nil {
		return errors.New("invalid parameter")
	}
	if f.rpcClients == nil {
		return errors.New("no any active rpc client")
	}
	//run on all rpc clients
	succeed := 0
	failed := 0
	for _, client := range f.rpcClients {
		if !client.IsActive() {
			failed++
			continue
		}
		bRet = client.DocSync(indexTag, docId, docJson)
		if bRet {
			succeed++
		}else{
			failed++
		}
	}
	if failed > 0 {
		return errors.New(fmt.Sprintf("failed:%v, succeed:%v", failed, succeed))
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