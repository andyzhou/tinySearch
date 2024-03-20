package rpc

import (
	"context"
	"errors"
	"github.com/andyzhou/tinysearch/define"
	search "github.com/andyzhou/tinysearch/pb"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"
)

/*
 * face for rpc client
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//doc sync request
type DocSyncReq struct {
	Tag string
	DocId string
	DocIds []string
	JsonByte []byte
	IsRemove bool
}

//face info
type Client struct {
	addr string
	isActive bool
	conn *grpc.ClientConn //rpc client connect
	client *search.SearchServiceClient //rpc client
	closeChan chan bool
	sync.RWMutex
}

//construct
func NewRpcClient(addr string) *Client {
	//self init
	this := &Client{
		addr:addr,
		closeChan: make(chan bool, 1),
	}

	//connect server
	this.connServer()

	//spawn main process
	go this.runMainProcess()
	return this
}

//quit
func (f *Client) Quit() {
	close(f.closeChan)
}

////////////////
//call api
////////////////

//create index
func (f *Client) IndexCreate(tag string) error {
	//check
	if tag == "" {
		return errors.New("invalid parameter")
	}

	//init real request
	realReq := &search.IndexCreateReq{
		Tag:tag,
	}

	//call doc query api
	_, err := (*f.client).IndexCreate(
		context.Background(),
		realReq,
	)
	return err
}

//query doc
func (f *Client) DocQuery(
		optKind int,
		tag string,
		optJson []byte,
	) ([]byte, error) {
	//check
	if tag == "" || optJson == nil {
		return nil, errors.New("invalid parameter")
	}

	//init real request
	realReq := &search.DocQueryReq{
		Kind:int32(optKind),
		Tag:tag,
		Json: optJson,
	}

	//call doc query api
	resp, err := (*f.client).DocQuery(
		context.Background(),
		realReq,
	)
	if err != nil {
		return nil, err
	}
	return resp.JsonByte, nil
}

//remove one or batch doc
func (f *Client) DocRemove(
		tag string,
		docIds ...string,
	) (bRet bool) {
	var (
		m any = nil
	)
	//basic check
	if tag == "" || docIds == nil || f.client == nil {
		bRet = false
		return
	}

	//try catch panic
	defer func() {
		if err := recover(); err != m {
			log.Println("RpcClient::DocRemove panic, err:", err)
			bRet = false
			return
		}
	}()

	//init request
	req := DocSyncReq{
		Tag:tag,
		DocIds:docIds,
		IsRemove:true,
	}
	//do sync req
	bRet = f.docSyncProcess(&req)
	return
}

//get one or batch doc
func (f *Client) DocGet(
		tag string,
		docIds ...string,
	) ([][]byte, error) {
	//check
	if tag == "" || docIds == nil {
		return nil, errors.New("invalid parameter")
	}

	//init real request
	realReq := &search.DocGetReq{
		Tag:tag,
		DocIds: docIds,
	}

	//call doc query api
	resp, err := (*f.client).DocGet(
		context.Background(),
		realReq,
	)
	if err != nil {
		return nil, err
	}
	return resp.JsonByte, nil
}

//sync doc
func (f *Client) DocSync(
		tag string,
		docId string,
		jsonByte []byte,
	) (bRet bool) {
	var (
		m any = nil
	)
	//basic check
	if tag == "" || jsonByte == nil || f.client == nil {
		bRet = false
		return
	}

	//try catch panic
	defer func() {
		if err := recover(); err != m {
			log.Println("RpcClient::DocSync panic, err:", err)
			bRet = false
			return
		}
	}()

	//init request
	req := DocSyncReq{
		Tag:tag,
		DocId:docId,
		JsonByte:jsonByte,
	}
	//do sync request
	bRet = f.docSyncProcess(&req)
	return
}

//check client is active or not
func (f *Client) IsActive() bool {
	return f.isActive
}

///////////////
//private func
///////////////

//run main process
func (f *Client) runMainProcess() {
	var (
		ticker = time.NewTicker(time.Second * define.ClientCheckTicker)
		m any = nil
	)

	//defer
	defer func() {
		if err := recover(); err != m {
			log.Println("RpcClient:mainProcess panic, err:", err)
		}
		ticker.Stop()
	}()

	//loop
	for {
		select {
		case <- ticker.C://check status
			{
				f.ping()
			}
		case <- f.closeChan:
			return
		}
	}
}

//doc sync into rpc server
func (f *Client) docSyncProcess(
		req *DocSyncReq,
	) bool {
	var (
		resp *search.DocSyncResp
		err error
	)

	if req == nil {
		return false
	}

	if req.IsRemove {
		//remove doc
		realReq := &search.DocRemoveReq{
			Tag:req.Tag,
			DocId:make([]string, 0),
		}
		if req.DocIds != nil && len(req.DocIds) > 0 {
			realReq.DocId = append(realReq.DocId, req.DocIds...)
		}else{
			realReq.DocId = append(realReq.DocId, req.DocId)
		}
		//call doc remove api
		resp, err = (*f.client).DocRemove(
			context.Background(),
			realReq,
		)
	}else{
		//add doc
		//init request
		realReq := &search.DocSyncReq{
			Tag:req.Tag,
			DocId:req.DocId,
			Json:req.JsonByte,
		}

		//call doc sync api
		resp, err = (*f.client).DocSync(
			context.Background(),
			realReq,
		)
	}

	if err != nil {
		log.Println("RpcClient::docSyncProcess failed, err:", err.Error())
		return false
	}
	return resp.Success
}

//ping server
func (f *Client) ping() bool {
	//check status
	isOk := f.checkStatus()
	if isOk {
		f.isActive = true
		return true
	}
	//try re-connect
	f.connServer()
	return true
}

//check server status
func (f *Client) checkStatus() bool {
	//check connect
	if f.conn == nil {
		return false
	}
	//get status
	state := f.conn.GetState().String()
	if state == "TRANSIENT_FAILURE" || state == "SHUTDOWN" {
		return false
	}
	return true
}

//connect rpc server
func (f *Client) connServer() error {
	//try connect
	f.isActive = false
	conn, err := grpc.Dial(f.addr, grpc.WithInsecure())
	if err != nil {
		log.Println("RpcClient::connServer failed, err:", err.Error())
		return err
	}

	//init rpc client
	client := search.NewSearchServiceClient(conn)
	if client == nil {
		return errors.New("init client failed")
	}

	//sync
	f.Lock()
	defer f.Unlock()
	f.conn = conn
	f.client = &client

	//ping server
	f.ping()
	return nil
}