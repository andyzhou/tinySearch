package rpc

import (
	"fmt"
	"github.com/andyzhou/tinysearch/iface"
	search "github.com/andyzhou/tinysearch/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

/*
 * face for rpc service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Service struct {
	addr string //rpc address
	docQueueMode bool
	queueWorkers int
	rpcCB *CB //rpc service
	manager iface.IManager //reference
	listener *net.Listener
	service *grpc.Server //rpc service
}

//construct
func NewRpcService(
		port int,
		manager iface.IManager,//reference
		docQueueMode bool,
		workers int,
	) *Service {
	//self init
	this := &Service{
		addr:fmt.Sprintf(":%d", port),
		manager:manager,
		docQueueMode: docQueueMode,
		queueWorkers: workers,
	}

	//create service
	this.createService()
	return this
}

//stop service
func (f *Service) Stop() {
	if f.service != nil {
		f.service.Stop()
	}
	if f.listener != nil {
		(*f.listener).Close()
	}
	if f.rpcCB != nil {
		f.rpcCB.Quit()
	}
}

/////////////////
//private func
/////////////////

//start service
func (f *Service) start() {
	//basic check
	if f.listener == nil || f.service == nil {
		return
	}

	//service rpc
	go f.beginService()
}

//begin rpc service
func (f *Service) beginService() {
	if f.listener == nil {
		return
	}

	//service listen
	err := f.service.Serve(*f.listener)
	if err != nil {
		log.Println("RpcService::beginService failed, err:", err.Error())
		panic(any(err))
	}
}

//create rpc service
func (f *Service) createService() {
	//listen tcp port
	listen, err := net.Listen("tcp", f.addr)
	if err != nil {
		log.Println("RpcService::createService failed, err:", err.Error())
		panic(any(err))
	}

	//create rpc server
	f.service = grpc.NewServer()

	//init rpc service
	f.rpcCB = NewRpcCB(f.manager, f.docQueueMode, f.queueWorkers)

	//register call back
	search.RegisterSearchServiceServer(
			f.service,
			f.rpcCB,
		)

	//sync listen
	f.listener = &listen

	//start service
	f.start()
}

