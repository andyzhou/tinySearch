package tinysearch

import (
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/face"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/rpc"
	"log"
	"sync"
)

/*
 * service api
 * - if none rpc mode, just opt base on service sub face
 */

var (
	_service *Service
	_serviceOnce sync.Once
)

//face info
type Service struct {
	manager iface.IManager
	rpcService iface.IRpcService
}

//get single instance
func GetService(rpcPort ...int) *Service {
	_serviceOnce.Do(func() {
		_service = NewService(rpcPort...)
	})
	return _service
}

//construct
func NewService(rpcPort ...int) *Service {
	//check and set rpc port
	//if rpc port > 0, will start rpc service
	rpcPortInt := 0
	if rpcPort != nil && len(rpcPort) > 0 {
		rpcPortInt = rpcPort[0]
	}
	return NewServiceWithPara(
				define.DataPathDefault,
				rpcPortInt,
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
	var (
		m any = nil
	)
	defer func() {
		if err := recover(); err != m {
			log.Printf("tinySearch.Service:Quit panic, err:%v", err)
		}
	}()
	f.manager.Quit()
	if f.rpcService != nil {
		f.rpcService.Stop()
	}
}

//set data path
func (f *Service) SetDataPath(path string) {
	f.manager.SetDataPath(path)
}

//set dict file path
func (f *Service) SetDictFile(filePath string, isForces ...bool) {
	var (
		isForce bool
	)
	if filePath == "" {
		filePath = define.SearchDictFileDefault
	}
	if isForces != nil && len(isForces) > 0 {
		isForce = isForces[0]
	}
	if !isForce {
		if f.manager.GetDictFile() != "" {
			//has set up, just return.
			return
		}
	}
	f.manager.SetDictFile(filePath)
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
