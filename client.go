package tinysearch

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/json"
	"github.com/andyzhou/tinysearch/lib"
	"github.com/andyzhou/tinysearch/rpc"
	"log"
	"sync"
)

/*
 * client api
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - used for rpc mode
 */

//query opt kind
const (
	QueryOptKindOfGen = iota
	QueryOptKindOfAgg
	QueryOptKindOfSuggest
)

const (
	DefaultWorkers = 9
)

//inter struct
type (
	syncDocReq struct {
		indexTag string
		docId string
		docJson []byte
	}
	removeDocReq struct {
		indexTag string
		docIds []string
	}
	getDocReq struct {
		indexTag string
		docIds []string
	}
	queryDocReq struct {
		queryKind int
		indexTag string
		optJson json.QueryOptJson
	}
)

//face info
type Client struct {
	rpcClients map[string]iface.IRpcClient //address -> rpcClient
	worker *lib.Worker
	workers int
	sync.RWMutex
}

//construct
func NewClient(workers ...int) *Client {
	var (
		workerNum int
	)
	//check workers
	if workers != nil && len(workers) > 0 {
		workerNum = workers[0]
	}
	if workerNum <= 0 {
		workerNum = DefaultWorkers
	}

	//self init
	this := &Client{
		rpcClients:make(map[string]iface.IRpcClient),
		worker: lib.NewWorker(),
		workers: workerNum,
	}
	this.interInit()
	return this
}

//quit
func (f *Client) Quit() {
	var (
		m any = nil
	)
	defer func() {
		if err := recover(); err != m {
			log.Printf("tinysearch.Client:Quit panic, err:%v", err)
		}
	}()
	if f.worker != nil {
		f.worker.Quit()
	}
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
	//check
	if indexTag == "" || optJson == nil {
		return nil, errors.New("invalid parameter")
	}

	//get rpc client
	client := f.getClient()
	if client == nil {
		return nil, errors.New("can't get active rpc client")
	}

	optJsonByte, err := optJson.Encode()
	if err != nil {
		return nil, err
	}

	//call api
	jsonByte, err := client.DocQuery(
		QueryOptKindOfSuggest,
		indexTag,
		optJsonByte,
	)
	if err != nil {
		return nil, err
	}

	//analyze result
	if jsonByte == nil {
		return nil, nil
	}

	//format result
	resultJson := json.NewSuggestsJson()
	err = resultJson.Decode(jsonByte)
	if err != nil {
		return nil, err
	}
	return resultJson, nil
}

//agg doc
func (f *Client) DocAgg(
		indexTag string,
		optJson *json.QueryOptJson,
	) (*json.AggregatesJson, error) {
	//check
	if indexTag == "" || optJson == nil {
		return nil, errors.New("invalid parameter")
	}

	//get rpc client
	client := f.getClient()
	if client == nil {
		return nil, errors.New("can't get active rpc client")
	}

	optJsonByte, err := optJson.Encode()
	if err != nil {
		return nil, err
	}

	//call api
	jsonByte, subErr := client.DocQuery(
									QueryOptKindOfAgg,
									indexTag,
									optJsonByte,
								)
	if subErr != nil {
		return nil, subErr
	}

	//analyze result
	if jsonByte == nil {
		return nil, nil
	}

	//format result
	resultJson := json.NewAggregatesJson()
	err = resultJson.Decode(jsonByte)
	if err != nil {
		return nil, err
	}
	return resultJson, nil
}

//query doc
func (f *Client) DocQuery(
		indexTag string,
		optJson *json.QueryOptJson,
	) (*json.SearchResultJson, error) {
	//check
	if indexTag == "" || optJson == nil {
		return nil, errors.New("invalid parameter")
	}
	if f.rpcClients == nil {
		return nil, errors.New("no any active rpc client")
	}
	//init request
	req := queryDocReq{
		indexTag: indexTag,
		optJson: *optJson,
	}

	resp, err := f.queryDoc(&req)
	return resp, err

	////send to worker queue
	//resp, err := f.worker.SendData(req, "", true)
	//resultObj, _ := resp.(*json.SearchResultJson)
	//return resultObj, err
}

//get doc
func (f *Client) DocGet(
		indexTag string,
		docIds ...string,
	) ([][]byte, error) {
	//check
	if indexTag == "" || docIds == nil {
		return nil, errors.New("invalid parameter")
	}
	if f.rpcClients == nil {
		return nil, errors.New("no any active rpc client")
	}

	//init request
	req := getDocReq{
		indexTag: indexTag,
		docIds: docIds,
	}

	//send worker queue
	resp, err := f.worker.SendData(req, docIds[0], true)
	if err != nil || resp == nil {
		return nil, err
	}
	respBytes, _ := resp.([][]byte)

	//call rpc api
	//return client.DocGet(indexTag, docIds...)
	return respBytes, err
}

//remove doc
func (f *Client) DocRemove(
		indexTag string,
		docIds ...string,
	) error {
	var (
		m any = nil
	)
	//check
	if indexTag == "" || docIds == nil {
		return errors.New("invalid parameter")
	}
	if f.rpcClients == nil {
		return errors.New("no any active rpc client")
	}

	//defer
	defer func() {
		if e := recover(); e != m {
			log.Printf("client.DocRemove panic, err:%v\n", e)
		}
	}()

	//init request
	req := removeDocReq{
		indexTag: indexTag,
		docIds: docIds,
	}

	//send to worker queue
	_, err := f.worker.SendData(req, docIds[0])
	return err
}

//add sync
//used for add, sync doc, run on all nodes
func (f *Client) DocSync(
		indexTag, docId string,
		docJson []byte,
	) error {
	var (
		m any = nil
	)
	//check
	if indexTag == "" || docId == "" || docJson == nil {
		return errors.New("invalid parameter")
	}
	if f.rpcClients == nil {
		return errors.New("no any active rpc client")
	}

	//defer
	defer func() {
		if e := recover(); e != m {
			log.Printf("client.DocSync panic, err:%v\n", e)
			return
		}
	}()

	//init request
	req := syncDocReq{
		indexTag: indexTag,
		docId: docId,
		docJson: docJson,
	}

	//send to worker queue
	_, err := f.worker.SendData(req, docId)
	return err
}

//create index
func (f *Client) CreateIndex(indexTag string) error {
	//check
	if indexTag == "" {
		return errors.New("invalid parameter")
	}
	//get rpc client
	client := f.getClient()
	if client == nil {
		return errors.New("can't get active rpc client")
	}
	//call rpc api
	err := client.IndexCreate(indexTag)
	return err
}

//add search service nodes
func (f *Client) AddNodes(nodes ... string) error {
	//check
	if nodes == nil || len(nodes) <= 0 {
		return errors.New("invalid parameter")
	}
	//check and init new rpc client
	f.Lock()
	defer f.Unlock()
	for _, node := range nodes {
		//check
		v, ok := f.rpcClients[node]
		if ok && v != nil {
			continue
		}
		//create new rpc client
		rpcClient := rpc.NewRpcClient(node)
		f.rpcClients[node] = rpcClient
	}
	return nil
}

//////////////
//private func
//////////////

//inter init
func (f *Client) interInit() {
	//init workers
	f.worker.SetCBForQueueOpt(f.cbForWorkerOpt)
	f.worker.CreateWorkers(f.workers)
}

//cb for worker queue opt
func (f *Client) cbForWorkerOpt(input interface{}) (interface{}, error) {
	//check
	if input == nil {
		return nil, errors.New("invalid parameter")
	}

	//do diff opt by data type
	switch dataType := input.(type) {
	case syncDocReq:
		{
			//sync doc req
			data, ok := input.(syncDocReq)
			if !ok || &data == nil {
				return nil, errors.New("invalid data type")
			}
			f.syncDoc(&data)
			break
		}
	case removeDocReq:
		{
			//remove doc req
			data, ok := input.(removeDocReq)
			if !ok || &data == nil {
				return nil, errors.New("invalid data type")
			}
			f.removeBatchDocByIds(&data)
			break
		}
	case getDocReq:
		{
			//get doc req
			data, ok := input.(getDocReq)
			if !ok || &data == nil {
				return nil, errors.New("invalid data type")
			}
			return f.getDoc(&data)
			break
		}
	case queryDocReq:
		{
			//query doc req
			data, ok := input.(queryDocReq)
			if !ok || &data == nil {
				return nil, errors.New("invalid data type")
			}
			return f.queryDoc(&data)
			break
		}
	default:
		{
			return nil, fmt.Errorf("invalid data type `%v`", dataType)
		}
	}
	return nil, nil
}

//query batch doc
func (f *Client) queryDoc(
		req *queryDocReq,
	) (*json.SearchResultJson, error) {
	//check
	if req == nil || req.indexTag == "" || &req.optJson == nil {
		return nil, errors.New("invalid parameter")
	}

	//get rpc client
	client := f.getClient()
	if client == nil {
		return nil, errors.New("can't get active rpc client")
	}

	optJsonByte, err := req.optJson.Encode()
	if err != nil {
		return nil, err
	}

	//call api
	jsonByte, subErr := client.DocQuery(
		QueryOptKindOfGen,
		req.indexTag,
		optJsonByte,
	)
	if subErr != nil {
		return nil, subErr
	}

	//analyze result
	if jsonByte == nil {
		return nil, nil
	}

	//format result
	resultJson := json.NewSearchResultJson()
	err = resultJson.Decode(jsonByte)
	return resultJson, err
}

//get one doc
func (f *Client) getDoc(req *getDocReq) ([][]byte, error) {
	//check
	if req == nil {
		return nil, errors.New("invalid parameter")
	}

	//get rpc client
	client := f.getClient()
	if client == nil {
		return nil, errors.New("can't get active rpc client")
	}

	//call rpc api
	resp, err := client.DocGet(req.indexTag, req.docIds...)
	return resp, err
}

//sync batch doc
func (f *Client) syncDoc(req *syncDocReq) bool {
	var (
		bRet bool
	)

	//check
	if req == nil || req.docJson == nil {
		return false
	}

	//run on all rpc clients
	succeed := 0
	failed := 0
	for _, client := range f.rpcClients {
		if !client.IsActive() {
			failed++
			continue
		}
		bRet = client.DocSync(req.indexTag, req.docId, req.docJson)
		if bRet {
			succeed++
		}else{
			failed++
		}
	}
	if failed > 0 {
		info := fmt.Sprintf("failed:%v, succeed:%v", failed, succeed)
		log.Printf("client:syncDoc, %v\n", info)
	}
	return true
}

//remove batch doc by ids
func (f *Client) removeBatchDocByIds(req *removeDocReq) bool {
	var (
		bRet bool
	)
	//check
	if req == nil || req.docIds == nil {
		return false
	}
	//run on all rpc clients
	succeed := 0
	failed := 0
	for _, client := range f.rpcClients {
		if !client.IsActive() {
			failed++
			continue
		}
		bRet = client.DocRemove(req.indexTag, req.docIds...)
		if bRet {
			succeed++
		}else{
			failed++
		}
	}
	if failed > 0 {
		info := fmt.Sprintf("failed:%v, succeed:%v", failed, succeed)
		log.Printf("client:removeBatchDocByIds, %v\n", info)
	}
	return true
}

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