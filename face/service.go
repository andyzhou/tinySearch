package face

import "github.com/andyzhou/tinySearch/iface"

/*
 * face for service
 * this is main entry
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Service struct {
	suggest iface.ISuggest
	agg iface.IAgg
	query iface.IQuery
	doc iface.IDoc
	manager iface.IManager
	rpc iface.IRpc
}

//construct
func NewService(port int) *Service {
	//self init
	this := &Service{
		manager:NewManager(),
		doc:NewDoc(),
		query:NewQuery(),
		agg:NewAgg(),
		suggest:NewSuggest(),
	}
	//init rpc
	this.rpc = NewRpc(port, this.manager)
	return this
}

//quit
func (f *Service) Quit() {
	f.rpc.Stop()
}

//get suggest face
func (f *Service) GetSuggest() iface.ISuggest {
	return f.suggest
}

//get agg face
func (f *Service) GetAgg() iface.IAgg {
	return f.agg
}

//get query face
func (f *Service) GetQuery() iface.IQuery {
	return f.query
}

//get doc face
func (f *Service) GetDoc() iface.IDoc {
	return f.doc
}

//get index face
func (f *Service) GetIndex(
					tag string,
				) iface.IIndex {
	return f.manager.GetIndex(tag)
}

//add index
func (f *Service) AddIndex(
					dir, tag string,
				) bool {
	return f.manager.AddIndex(dir, tag)
}

//add rpc node
func (f *Service) AddNode(
					addr string,
				) bool {
	return f.manager.AddNode(addr)
}
