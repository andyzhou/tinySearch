package face

import (
	"context"
	"errors"
	"github.com/andyzhou/tinySearch/define"
	search "github.com/andyzhou/tinySearch/pb"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"
)

/*
 * face for rpc client
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
type RpcClient struct {
	addr string
	isActive bool
	conn *grpc.ClientConn //rpc client connect
	client *search.SearchServiceClient //rpc client
	docSyncChan chan DocSyncReq
	closeChan chan bool
	sync.RWMutex
}

//construct
func NewRpcClient(addr string) *RpcClient {
	//self init
	this := &RpcClient{
		addr:addr,
		docSyncChan:make(chan DocSyncReq, define.ReqChanSize),
		closeChan:make(chan bool, 1),
	}

	//try connect server
	this.connServer()

	//spawn main process
	go this.runMainProcess()

	return this
}

//quit
func (f *RpcClient) Quit() {
	f.closeChan <- true
}

//call api
func (f *RpcClient) DocQuery(optKind int, tag string, optJson []byte) ([]byte, error) {
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

func (f *RpcClient) DocRemove(
					tag string,
					docIds ...string,
				) (bRet bool) {
	//basic check
	if tag == "" || docIds == nil || f.client == nil {
		bRet = false
		return
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
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

	//send to chan
	f.docSyncChan <- req
	bRet = true
	return
}

func (f *RpcClient) DocSync(
					tag string,
					docId string,
					jsonByte []byte,
				) (bRet bool) {
	//basic check
	if tag == "" || jsonByte == nil || f.client == nil {
		bRet = false
		return
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
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

	//send to chan
	f.docSyncChan <- req
	bRet = true
	return
}

//check client is active or not
func (f *RpcClient) IsActive() bool {
	return f.isActive
}

///////////////
//private func
///////////////

//run main process
func (f *RpcClient) runMainProcess() {
	var (
		ticker = time.NewTicker(time.Second * define.ClientCheckTicker)
		req DocSyncReq
		isOk, needQuit bool
	)

	//loop
	for {
		if needQuit {
			break
		}
		select {
		case req, isOk = <- f.docSyncChan://doc sync req
			if isOk {
				f.docSyncProcess(&req)
			}
		case <- ticker.C://check status
			{
				f.ping()
			}
		case <- f.closeChan:
			needQuit = true
		}
	}

	//close chan
	close(f.docSyncChan)
	close(f.closeChan)
}

//doc sync into rpc server
func (f *RpcClient) docSyncProcess(
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
			DocId:[]string{
				req.DocId,
			},
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
func (f *RpcClient) ping() bool {
	//check status
	isOk := f.checkStatus()
	if isOk {
		f.isActive = false
		return true
	}
	//try re connect
	f.connServer()
	return true
}

//check server status
func (f *RpcClient) checkStatus() bool {
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
func (f *RpcClient) connServer() bool {
	//try connect
	conn, err := grpc.Dial(f.addr, grpc.WithInsecure())
	if err != nil {
		log.Println("RpcClient::connServer failed, err:", err.Error())
		return false
	}

	//init rpc client
	client := search.NewSearchServiceClient(conn)
	if client == nil {
		return false
	}

	//sync
	f.Lock()
	defer f.Unlock()
	f.conn = conn
	f.client = &client

	//ping server
	f.ping()

	return true
}