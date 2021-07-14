package iface

/*
 * interface for rpc client
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IClient interface {
	Quit()
	DocQuery(optKind int, tag string, optJson []byte) ([][]byte, int32, error)
	DocRemove(tag string, docIds []string) bool
	DocSync(tag, docId string, jsonByte []byte) bool
	IsActive() bool
}

