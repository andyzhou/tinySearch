package tinySearch

import (
	"github.com/andyzhou/tinySearch/define"
	"github.com/andyzhou/tinySearch/face"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/rpc"
	"log"
)

/*
 * service api
 */

//face info
type Service struct {
	manager iface.IManager
	rpcService iface.IRpcService
}

//construct
//if rpc port > 0, will start rpc service
func NewService(rpcPort ...int) *Service {
	rpcPortInt := 0
	if rpcPort != nil && len(rpcPort) > 0 {
		rpcPortInt = rpcPort[0]
	}
	return NewServiceWithPara(
				define.DataPathDefault,
				rpcPortInt,
				define.SearchDictFileDefault,
			)
}

//construct with parameter
func NewServiceWithPara(
			dataPath string,
			rpcPort int,
			dictFile ...string,
		) *Service {
	if dataPath == "" {
		dataPath = define.DataPathDefault
	}
	//self init
	this := &Service{
		manager: face.NewManager(dataPath, dictFile...),
	}
	//init rpc if rpc port > 0
	if rpcPort > 0 {
		this.rpcService = rpc.NewRpcService(rpcPort, this.manager)
	}
	return this
}

//quit
func (f *Service) Quit() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Service:Quit panic, err:", err)
		}
	}()
	f.manager.Quit()
	if f.rpcService != nil {
		f.rpcService.Stop()
	}
}

//get suggest face
func (f *Service) GetSuggest() iface.ISuggest {
	return f.manager.GetSuggest()
}

//get agg face
func (f *Service) GetAgg() iface.IAgg {
	return f.manager.GetAgg()
}

//get query face
func (f *Service) GetQuery() iface.IQuery {
	return f.manager.GetQuery()
}

//get doc face
func (f *Service) GetDoc() iface.IDoc {
	return f.manager.GetDoc()
}

//get index face
func (f *Service) GetIndex(tag string) iface.IIndex {
	return f.manager.GetIndex(tag)
}

//add index
func (f *Service) AddIndex(tag string) error {
	return f.manager.AddIndex(tag)
}
