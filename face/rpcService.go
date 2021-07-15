package face

import (
	"fmt"
	"github.com/andyzhou/tinySearch/iface"
	search "github.com/andyzhou/tinySearch/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

/*
 * face for rpc service
 */

//face info
type RpcService struct {
	addr string //rpc address
	manager iface.IManager //reference
	listener *net.Listener
	service *grpc.Server //rpc service
}

//construct
func NewRpcService(
			port int,
			manager iface.IManager,//reference
		) *RpcService {
	//self init
	this := &RpcService{
		addr:fmt.Sprintf(":%d", port),
		manager:manager,
	}
	//create service
	this.createService()
	return this
}

//stop service
func (f *RpcService) Stop() {
	if f.service != nil {
		f.service.Stop()
	}
	if f.listener != nil {
		(*f.listener).Close()
	}
}

/////////////////
//private func
/////////////////

//start service
func (f *RpcService) start() {
	//basic check
	if f.listener == nil || f.service == nil {
		return
	}

	//service rpc
	go f.beginService()
}

//begin rpc service
func (f *RpcService) beginService() {
	if f.listener == nil {
		return
	}

	//service listen
	err := f.service.Serve(*f.listener)
	if err != nil {
		log.Println("RpcService::beginService failed, err:", err.Error())
		panic(err)
	}
}

//create rpc service
func (f *RpcService) createService() {
	//listen tcp port
	listen, err := net.Listen("tcp", f.addr)
	if err != nil {
		log.Println("RpcService::createService failed, err:", err.Error())
		panic(err.Error())
	}

	//create rpc server
	f.service = grpc.NewServer()

	//register call back
	search.RegisterSearchServiceServer(
			f.service,
			NewRpcCB(f.manager),
		)
	f.listener = &listen

	//start service
	f.start()
}

