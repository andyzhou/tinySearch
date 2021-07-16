package face

import (
	"context"
	"errors"
	"fmt"
	"github.com/andyzhou/tinySearch/define"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	search "github.com/andyzhou/tinySearch/pb"
)

/*
 * rpc call back for rpc service
 */

//face info
type RpcCB struct {
	manager iface.IManager //manager reference
	json.BaseJson
}

//construct
func NewRpcCB(
			manager iface.IManager,
		) *RpcCB {
	//self init
	this := &RpcCB{
		manager:manager,
	}
	return this
}

/////////////////////////////
//call backs for rpc service
/////////////////////////////

//doc query
func (f *RpcCB) DocQuery(
					ctx context.Context,
					in *search.DocQueryReq,
				) (*search.DocQueryResp, error) {
	var (
		tip string
		jsonByte []byte
		err error
	)

	//check input
	if in == nil {
		return nil, errors.New("invalid parameter")
	}

	//decode query opt json
	queryOptJson := json.NewQueryOptJson()
	bRet := queryOptJson.Decode(in.Json)
	if !bRet {
		tip = fmt.Sprintf("invalid query opt json")
		return nil, errors.New(tip)
	}

	//get index
	index := f.manager.GetIndex(in.Tag)
	if index == nil {
		tip = fmt.Sprintf("can't get index by tag of %s", in.Tag)
		return nil, errors.New(tip)
	}

	//format result
	resp := &search.DocQueryResp{
		JsonByte: make([]byte, 0),
	}

	//get key data
	optKind := in.Kind

	//do diff opt by kind
	switch optKind {
	case define.QueryOptKindOfAgg:
		{
			jsonByte, err = f.aggDocQuery(index, queryOptJson)
		}
	case define.QueryOptKindOfSuggest:
		{
			jsonByte, err = f.suggestDocQuery(index, queryOptJson)
		}
	case define.QueryOptKindOfGen:
		fallthrough
	default:
		{
			jsonByte, err = f.genDocQuery(index, queryOptJson)
		}
	}

	//format response
	if err != nil {
		return nil, err
	}
	resp.Success = true
	resp.JsonByte = jsonByte

	return resp, nil
}

//doc remove
func (f *RpcCB) DocRemove(
					ctx context.Context,
					in *search.DocRemoveReq,
				) (*search.DocSyncResp, error) {
	var (
		tip string
	)

	//check input value
	if in == nil {
		return nil, errors.New("invalid parameter")
	}

	//get index
	index := f.manager.GetIndex(in.Tag)
	if index == nil {
		tip = fmt.Sprintf("can't get index by tag of %s", in.Tag)
		return nil, errors.New(tip)
	}

	//remove from local index
	indexer := index.GetIndex()
	for _, docId := range in.DocId {
		err := indexer.Delete(docId)
		if err != nil {
			return nil, errors.New(err.Error())
		}
	}
	//format result
	result := &search.DocSyncResp{
		Success:true,
	}
	return result, nil
}

//doc sync
func (f *RpcCB) DocSync(
					ctx context.Context,
					in *search.DocSyncReq,
				) (*search.DocSyncResp, error) {
	var (
		tip string
	)

	//check input value
	if in == nil {
		return nil, errors.New("invalid parameter")
	}

	//get index
	index := f.manager.GetIndex(in.Tag)
	if index == nil {
		tip = fmt.Sprintf("can't get index by tag of %s", in.Tag)
		return nil, errors.New(tip)
	}

	//decode json byte
	kvMap := make(map[string]interface{})
	bRet := f.BaseJson.DecodeSimple(in.Json, kvMap)
	if !bRet {
		return nil, errors.New("decode json byte failed")
	}

	//add into local index
	indexer := index.GetIndex()
	err := indexer.Index(in.DocId, kvMap)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	//format result
	result := &search.DocSyncResp{
		Success:true,
	}
	return result, nil
}

/////////////////
//private func
/////////////////

//suggest query
func (f *RpcCB) suggestDocQuery(
					index iface.IIndex,
					queryOptJson *json.QueryOptJson,
				) ([]byte, error) {
	//suggest doc
	suggest := f.manager.GetSuggest()
	suggestOptJson := json.NewSuggestOptJson()
	suggestOptJson.Key = queryOptJson.Key
	suggestOptJson.Size = queryOptJson.PageSize
	suggestListJson, err := suggest.GetSuggest(suggestOptJson)
	if err != nil {
		return nil, err
	}
	return suggestListJson.Encode(), nil
}

//agg query
func (f *RpcCB) aggDocQuery(
					index iface.IIndex,
					queryOptJson *json.QueryOptJson,
				) ([]byte, error) {
	//agg doc
	agg := f.manager.GetAgg()
	aggListJson, err := agg.GetAggList(index, queryOptJson)
	if err != nil {
		return nil, err
	}
	return aggListJson.Encode(), nil
}

//general query
func (f *RpcCB) genDocQuery(
					index iface.IIndex,
					queryOptJson *json.QueryOptJson,
				 ) ([]byte, error) {
	//query doc
	query := f.manager.GetQuery()
	resultsJson, err := query.Query(index, queryOptJson)
	if err != nil {
		return nil, err
	}
	return resultsJson.Encode(), nil
}