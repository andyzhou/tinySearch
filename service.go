package tinySearch

import (
	"github.com/andyzhou/tinySearch/face"
	"github.com/andyzhou/tinySearch/iface"
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
func NewService(dataPath string, rpcPort int) *Service {
	//self init
	this := &Service{
		manager: face.NewManager(dataPath),
	}
	//init rpc
	this.rpcService = face.NewRpcService(rpcPort, this.manager)
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
	f.rpcService.Stop()
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
func (f *Service) GetIndex(
					tag string,
				) iface.IIndex {
	return f.manager.GetIndex(tag)
}

//add index
func (f *Service) AddIndex(
					tag string,
				) bool {
	return f.manager.AddIndex(tag)
}