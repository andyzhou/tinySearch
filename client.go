package tinySearch

import (
	"errors"
	"github.com/andyzhou/tinySearch/face"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
)

/*
 * client api
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//query opt kind
const (
	QueryOptKindOfGen = iota
	QueryOptKindOfAgg
	QueryOptKindOfSuggest
)

//face info
type Client struct {
	manager iface.IManager
}

//construct
func NewClient() *Client {
	//self init
	self := &Client{
		manager: face.NewManager(""),
	}
	return self
}

//quit
func (f *Client) Quit() {
	f.manager.Quit()
}

//suggest doc
func (f *Client) DocSuggest(
					indexTag string,
					optJson *json.QueryOptJson,
				) (*json.SuggestsJson, error) {
	var (
		bRet bool
	)

	//get client
	client := f.manager.GetClient()
	if client == nil {
		return nil, errors.New("can't get client")
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

	//get client
	client := f.manager.GetClient()
	if client == nil {
		return nil, errors.New("can't get client")
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

	//get client
	client := f.manager.GetClient()
	if client == nil {
		return nil, errors.New("can't get client")
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
func (f *Client) DocRemove(tag, docId string) error {
	//remove doc from relate nodes pass manager
	err := f.manager.DocRemove(tag, docId)
	return err
}

//sync doc
func (f *Client) DocSync(tag, docId string, docJson []byte) error {
	//sync doc to relate nodes pass manager
	err := f.manager.DocSync(tag, docId, docJson)
	return err
}

//add search service nodes
func (f *Client) AddNodes(nodes ... string) bool {
	//check
	if nodes == nil || len(nodes) <= 0 {
		return false
	}

	//add into manager
	bRet := f.manager.AddNode(nodes...)
	return bRet
}