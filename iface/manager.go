package iface

/*
 * interface for inter manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IManager interface {
	Quit()

	//for doc sync and remove
	DocRemove(tag, docId string) bool
	DocSync(tag, docId string, jsonByte []byte) bool

	//for rpc node
	RemoveNode(addr string) bool
	AddNode(addr string) bool

	//for index
	RemoveIndex(tag string) bool
	GetIndex(tag string) IIndex
	AddIndex(dir, tag string) bool
}