package iface

/*
 * interface for service, main entry
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IService interface {
	Quit()
	GetSuggest() ISuggest
	GetAgg() IAgg
	GetQuery() IQuery
	GetDoc() IDoc
	GetIndex(tag string) IIndex
	AddIndex(dir, tag string) bool
	AddNode(addr string) bool
}


