package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/json"
	"github.com/andyzhou/tinysearch/lib"
	search "github.com/andyzhou/tinysearch/pb"
)

/*
 * rpc call back for rpc service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//inter macro define
const (
)

//inter type
type (
	GetDocQueueReq struct {
		In search.DocGetReq
		Out chan search.DocGetResp
	}
	AddDocQueueReq struct{
		In search.DocSyncReq
		Out chan search.DocSyncResp
	}
	RemoveDocQueueReq struct {
		In search.DocRemoveReq
		Out chan search.DocSyncResp
	}
)

//face info
type CB struct {
	docQueueMode bool
	worker *lib.Worker
	manager iface.IManager //manager reference
	json.BaseJson
}

//construct
func NewRpcCB(
		manager iface.IManager,
		docQueueMode bool,
		queueWorkers int,
	) *CB {
	//self init
	this := &CB{
		manager:manager,
		docQueueMode: docQueueMode,
		worker: lib.NewWorker(),
	}

	//check and init inter add doc queue
	if docQueueMode {
		//set consumer
		this.worker.SetCBForQueueOpt(this.cbForQueue)

		//create batch son workers
		if queueWorkers <= 0 {
			queueWorkers = lib.DefaultQueueSize
		}
		this.worker.CreateWorkers(queueWorkers)
	}
	return this
}

//quit
func (f *CB) Quit() {
	if f.worker != nil {
		f.worker.Quit()
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

//doc get
func (f *CB) DocGet(
		ctx context.Context,
		in *search.DocGetReq,
	) (*search.DocGetResp, error) {
	//check input value
	if in == nil {
		return nil, errors.New("invalid parameter")
	}

	//check queue mode
	if !in.UseQueue {
		//just call low level api
		resp, err := f.lowLevelGetDoc(in)
		return resp, err
	}

	//format and send to queue
	queueReq := GetDocQueueReq{
		In: *in,
		Out: make(chan search.DocGetResp, 1),
	}

	//send to inter queue list
	f.worker.SendData(queueReq, in.DocIds[0])

	//wait response
	resp, ok := <- queueReq.Out
	if !ok || &resp == nil {
		return nil, errors.New("can't get queue response")
	}
	return &resp, nil
}

//doc remove
func (f *CB) DocRemove(
		ctx context.Context,
		in *search.DocRemoveReq,
	) (*search.DocSyncResp, error) {
	//check input value
	if in == nil {
		return nil, errors.New("invalid parameter")
	}

	//check queue mode
	if !f.docQueueMode {
		//just call low level api
		resp, err := f.lowLevelRemoveDoc(in)
		return resp, err
	}

	//format and send to queue
	queueReq := RemoveDocQueueReq{
		In: *in,
		Out: make(chan search.DocSyncResp, 1),
	}

	//send to inter queue list
	f.worker.SendData(queueReq, in.DocId[0])

	//wait response
	resp, ok := <- queueReq.Out
	if !ok || &resp == nil {
		return nil, errors.New("can't get queue response")
	}
	return &resp, nil
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
	if !f.docQueueMode {
		//just call low level api
		resp, err := f.lowLevelAddDoc(in)
		return resp, err
	}

	//format and send to queue
	queueReq := AddDocQueueReq{
		In: *in,
		Out: make(chan search.DocSyncResp, 1),
	}

	//send to inter queue list
	f.worker.SendData(queueReq, in.DocId)

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

//low level get doc
func (f *CB) lowLevelGetDoc(
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

//low level remove doc
func (f *CB) lowLevelRemoveDoc(
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

//low level add doc
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


//get doc request opt
func (f *CB) getDocReqOpt(req *GetDocQueueReq) error {
	//check
	if req == nil || &req.In == nil {
		return errors.New("invalid parameter")
	}

	//call api func
	resp, err := f.lowLevelGetDoc(&req.In)
	if resp == nil {
		resp = &search.DocGetResp{}
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

//remove doc request opt
func (f *CB) removeDocReqOpt(req *RemoveDocQueueReq) error {
	//check
	if req == nil || &req.In == nil {
		return errors.New("invalid parameter")
	}

	//call api func
	resp, err := f.lowLevelRemoveDoc(&req.In)
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

//cb for queue opt
func (f *CB) cbForQueue(input interface{}) (interface{}, error) {
	var (
		err error
	)
	//check
	if input == nil {
		return nil, errors.New("invalid parameter")
	}

	//do diff opt by data type
	switch dataType := input.(type) {
	case GetDocQueueReq:
		{
			//get doc opt
			req, ok := input.(GetDocQueueReq)
			if !ok || &req == nil {
				return nil, errors.New("invalid input type")
			}
			err = f.getDocReqOpt(&req)
		}
	case AddDocQueueReq:
		{
			//add doc opt
			req, ok := input.(AddDocQueueReq)
			if !ok || &req == nil {
				return nil, errors.New("invalid input type")
			}
			err = f.addDocReqOpt(&req)
		}
	case RemoveDocQueueReq:
		{
			//remove doc opt
			req, ok := input.(RemoveDocQueueReq)
			if !ok || &req == nil {
				return nil, errors.New("invalid input type")
			}
			err = f.removeDocReqOpt(&req)
		}
	default:
		{
			err = fmt.Errorf("invalid data type of `%v`", dataType)
		}
	}
	return nil, err
}