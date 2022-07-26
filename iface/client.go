package iface

/*
 * interface for rpc client
 */

type IRpcClient interface {
	Quit()
	DocQuery(optKind int, tag string, optJson []byte) ([]byte, error)
	DocRemove(tag string, docIds ...string) bool
	DocGet(tag string, docIds ...string) ([][]byte, error)
	DocSync(tag, docId string, jsonByte []byte) bool
	IndexCreate(tag string) error
	IsActive() bool
}

