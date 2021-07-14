package iface

/*
 * interface for inter manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IManager interface {
	Quit()

	//for doc sync and remove
	DocsRemove(tag string, docIds []string) bool
	DocRemove(tag, docId string) error
	DocSync(tag, docId string, jsonByte []byte) error

	//for rpc node
	RemoveNode(addr string) bool
	AddNode(addr ...string) bool

	//for index
	RemoveIndex(tag string) bool
	GetIndex(tag string) IIndex
	AddIndex(tag string) bool

	//get rand client
	GetClient() IClient
}