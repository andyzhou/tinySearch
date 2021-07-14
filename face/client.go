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
	docSyncChan chan DocSyncReq
	closeChan chan bool
	sync.RWMutex
}

//construct
func NewClient(addr string) *Client {
	//self init
	this := &Client{
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
func (f *Client) Quit() {
	f.closeChan <- true
}

//call api
func (f *Client) DocQuery(optKind int, tag string, optJson []byte) ([][]byte, int32, error) {
	//check
	if tag == "" || optJson == nil {
		return nil, 0, errors.New("invalid parameter")
	}

	//init real request
	realReq := &search.DocQueryReq{
		Tag:tag,
		Json: optJson,
	}

	//call doc query api
	resp, err := (*f.client).DocQuery(
		context.Background(),
		realReq,
	)
	if err != nil {
		return nil, 0, err
	}
	return resp.RecList, resp.Total, nil
}

func (f *Client) DocRemove(
					tag string,
					docIds []string,
				) (bRet bool) {
	//basic check
	if tag == "" || docIds == nil || f.client == nil {
		bRet = false
		return
	}

	//try catch panic
	defer func() {
		if err := recover(); err != nil {
			log.Println("Client::DocRemove panic, err:", err)
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

func (f *Client) DocSync(
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
			log.Println("Client::DocSync panic, err:", err)
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
		log.Println("Client::docSyncProcess failed, err:", err.Error())
		return false
	}

	return resp.Success
}

//ping server
func (f *Client) ping() bool {
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
func (f *Client) connServer() bool {
	//try connect
	conn, err := grpc.Dial(f.addr, grpc.WithInsecure())
	if err != nil {
		log.Println("Client::connServer failed, err:", err.Error())
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
	f.isActive = true

	return true
}