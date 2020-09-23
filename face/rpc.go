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
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Rpc struct {
	addr string //rpc address
	manager iface.IManager
	listener *net.Listener
	service *grpc.Server //rpc service
}

//construct
func NewRpc(
			port int,
			manager iface.IManager,//reference
		) *Rpc {
	//self init
	this := &Rpc{
		addr:fmt.Sprintf(":%d", port),
		manager:manager,
	}
	//create service
	this.createService()
	return this
}

//stop service
func (f *Rpc) Stop() {
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
func (f *Rpc) start() {
	//basic check
	if f.listener == nil || f.service == nil {
		return
	}

	//service rpc
	go f.beginService()
}

//begin rpc service
func (f *Rpc) beginService() {
	if f.listener == nil {
		return
	}

	//service listen
	err := f.service.Serve(*f.listener)
	if err != nil {
		log.Println("Rpc::beginService failed, err:", err.Error())
		panic(err)
	}
}

//create rpc service
func (f *Rpc) createService() {
	//listen tcp port
	listen, err := net.Listen("tcp", f.addr)
	if err != nil {
		log.Println("Rpc::createService failed, err:", err.Error())
		panic(err.Error())
	}

	//create rpc server
	f.service = grpc.NewServer()

	//register call back
	search.RegisterSearchServiceServer(
			f.service,
			NewIRpcCB(f.manager),
		)
	f.listener = &listen

	//start service
	f.start()
}

