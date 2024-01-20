package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/json"
	search "github.com/andyzhou/tinysearch/pb"
	"log"
	"runtime/debug"
)

/*
 * rpc call back for rpc service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter macro define
const (
	AddDocQueueSize = 1024
)

//inter type
type (
	AddDocQueueReq struct{
		In search.DocSyncReq
		Out chan search.DocSyncResp //sync response chan
	}
)

//face info
type CB struct {
	addDocQueueMode bool
	manager iface.IManager //manager reference
	addDocQueueSize int
	addDocQueue chan AddDocQueueReq //add doc queue
	addDocCloseChan chan bool
	json.BaseJson
}

//construct
func NewRpcCB(
			manager iface.IManager,
			addDocQueueMode bool,
			addDocQueueSizes ...int,
		) *CB {
	var (
		addDocQueueSize int
	)
	//detect queue size
	if addDocQueueSizes != nil && len(addDocQueueSizes) > 0 {
		addDocQueueSize = addDocQueueSizes[0]
	}
	if addDocQueueSize <= 0 {
		addDocQueueSize = AddDocQueueSize
	}

	//self init
	this := &CB{
		manager:manager,
		addDocQueueMode: addDocQueueMode,
		addDocQueueSize: addDocQueueSize,
	}

	//check and init inter add doc queue
	if addDocQueueMode {
		this.addDocQueue = make(chan AddDocQueueReq, addDocQueueSize)
		this.addDocCloseChan = make(chan bool, 1)

		//spawn son processor
		go this.runAddDocQueueProcessor()
	}
	return this
}

//quit
func (f *CB) Quit() {
	if f.addDocCloseChan != nil && f.addDocQueue != nil {
		f.addDocCloseChan <- true
	}
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
		hitDocByte, subErr := hitDoc.Encode()
		if subErr != nil || hitDocByte == nil {
			continue
		}
		result.JsonByte = append(result.JsonByte, hitDocByte)
	}
	return result, nil
}

//doc sync
func (f *CB) DocSync(
					ctx context.Context,
					in *search.DocSyncReq,
				) (*search.DocSyncResp, error) {
	//check input value
	if in == nil {
		return nil, errors.New("invalid parameter")
	}

	//check queue mode
	if !f.addDocQueueMode {
		//just call low level api
		resp, err := f.lowLevelAddDoc(in)
		return resp, err
	}

	//check inter queue
	if f.addDocQueue == nil || len(f.addDocQueue) >= AddDocQueueSize {
		return nil, errors.New("inter add doc queue is nil or full")
	}

	//format and send to queue
	queueReq := AddDocQueueReq{
		In: *in,
		Out: make(chan search.DocSyncResp, 1),
	}

	//send to inter queue
	f.addDocQueue <- queueReq

	//wait response
	resp, ok := <- queueReq.Out
	if !ok || &resp == nil {
		return nil, errors.New("can't get queue response")
	}
	return &resp, nil
}

/////////////////
//private func
/////////////////

//low level add doc function
func (f *CB) lowLevelAddDoc(
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

	//check and call add doc hook
	doc := f.manager.GetDoc()
	docAddHook := doc.GetHoodForAddDoc()
	if docAddHook != nil {
		//has register hook, call it
		subErr := docAddHook(in.Json)
		if subErr != nil {
			return nil, subErr
		}
	}

	//decode json byte as general kv map
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

//suggest query
func (f *CB) suggestDocQuery(
					index iface.IIndex,
					queryOptJson *json.QueryOptJson,
				) ([]byte, error) {
	//suggest doc
	suggest := f.manager.GetSuggest()
	suggestOptJson := json.NewSuggestOptJson()
	suggestOptJson.QueryKind = queryOptJson.QueryKind
	suggestOptJson.Key = queryOptJson.Key
	suggestOptJson.IndexTag = queryOptJson.SuggestTag
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

//add doc request opt
func (f *CB) addDocReqOpt(req *AddDocQueueReq) error {
	//check
	if req == nil || &req.In == nil {
		return errors.New("invalid parameter")
	}

	//call api func
	resp, err := f.lowLevelAddDoc(&req.In)
	if resp == nil {
		resp = &search.DocSyncResp{}
	}
	if err != nil {
		resp.ErrMsg = err.Error()
	}else{
		resp.Success = true
	}

	//send response
	defer func() {
		//send to out chan
		if req.Out != nil {
			req.Out <- *resp
		}
	}()
	return err
}

//add doc queue processor
func (f *CB) runAddDocQueueProcessor() {
	var (
		req AddDocQueueReq
		isOk bool
		m any = nil
	)

	//defer
	defer func() {
		if err := recover(); err != m {
			log.Printf("tinysearch.rpcCB.runAddDocQueueProcessor panic, err:%v\n", err)
			log.Printf("tinysearch.rpcCB, track:%v\n", string(debug.Stack()))
		}
	}()

	//loop
	for {
		select {
		case req, isOk = <- f.addDocQueue:
			if isOk && &req != nil {
				//add doc opt
				f.addDocReqOpt(&req)
			}
		}
	}
}