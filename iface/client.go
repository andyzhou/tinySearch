package iface

/*
 * interface for rpc client
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IClient interface {
	Quit()
	DocRemove(tag, docId string) bool
	DocSync(tag, docId string, jsonByte []byte) bool
}

