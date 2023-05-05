package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/json"
	search "github.com/andyzhou/tinysearch/pb"
)

/*
 * rpc call back for rpc service
 */

//face info
type CB struct {
	manager iface.IManager //manager reference
	json.BaseJson
}

//construct
func NewRpcCB(
			manager iface.IManager,
		) *CB {
	//self init
	this := &CB{
		manager:manager,
	}
	return this
}

/////////////////////////////
//call backs for rpc service
/////////////////////////////

//index create
func (f *CB) IndexCreate(
					ctx context.Context,
					in *search.IndexCreateReq,
				) (*search.IndexCreateResp, error) {
	//check input
	if in == nil {
		return nil, errors.New("invalid parameter")
	}

	//check index
	index := f.manager.GetIndex(in.Tag)
	if index != nil {
		return nil, errors.New("index tag has exists")
	}

	//create new index
	err := f.manager.AddIndex(in.Tag)
	if err != nil {
		return nil, err
	}

	//format response
	resp := &search.IndexCreateResp{}
	resp.Success = true
	return resp, nil
}

//doc query
func (f *CB) DocQuery(
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
	err = queryOptJson.Decode(in.Json)
	if err != nil {
		return nil, err
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
func (f *CB) DocRemove(
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

//doc get
func (f *CB) DocGet(
				ctx context.Context,
				in *search.DocGetReq,
			) (*search.DocGetResp, error) {
	var (
		tip string
	)
	//check input value
	if in == nil || in.DocIds == nil {
		return nil, errors.New("invalid parameter")
	}

	//get index
	index := f.manager.GetIndex(in.Tag)
	if index == nil {
		tip = fmt.Sprintf("can't get index by tag of %s", in.Tag)
		return nil, errors.New(tip)
	}

	//get doc face
	doc := f.manager.GetDoc()

	//get batch docs
	hitDocs, err := doc.GetDocs(index, in.DocIds...)
	if err != nil {
		return nil, err
	}
	if hitDocs == nil {
		return nil, errors.New("no any records")
	}

	//format result
	result := &search.DocGetResp{
		Success:true,
		JsonByte:make([][]byte, 0),
	}
	for _, hitDoc := range hitDocs {
		result.JsonByte = append(result.JsonByte, hitDoc.OrgJson)
	}
	return result, nil
}

//doc sync
func (f *CB) DocSync(
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
	err := f.BaseJson.DecodeSimple(in.Json, kvMap)
	if err != nil {
		return nil, err
	}

	//add into local index
	indexer := index.GetIndex()
	err = indexer.Index(in.DocId, kvMap)
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
func (f *CB) suggestDocQuery(
					index iface.IIndex,
					queryOptJson *json.QueryOptJson,
				) ([]byte, error) {
	//suggest doc
	suggest := f.manager.GetSuggest()
	suggestOptJson := json.NewSuggestOptJson()
	suggestOptJson.Key = queryOptJson.Key
	suggestOptJson.Page = queryOptJson.Page
	suggestOptJson.PageSize = queryOptJson.PageSize
	suggestListJson, err := suggest.GetSuggest(suggestOptJson)
	if err != nil {
		return nil, err
	}
	if suggestListJson == nil {
		return nil, nil
	}
	return suggestListJson.Encode()
}

//agg query
func (f *CB) aggDocQuery(
					index iface.IIndex,
					queryOptJson *json.QueryOptJson,
				) ([]byte, error) {
	//agg doc
	agg := f.manager.GetAgg()
	aggListJson, err := agg.GetAggList(index, queryOptJson)
	if err != nil {
		return nil, err
	}
	if aggListJson == nil {
		return nil, nil
	}
	return aggListJson.Encode()
}

//general query
func (f *CB) genDocQuery(
					index iface.IIndex,
					queryOptJson *json.QueryOptJson,
				 ) ([]byte, error) {
	//query doc
	query := f.manager.GetQuery()
	resultsJson, err := query.Query(index, queryOptJson)
	if err != nil {
		return nil, err
	}
	if resultsJson == nil {
		return nil, nil
	}
	return resultsJson.Encode()
}