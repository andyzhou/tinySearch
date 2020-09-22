package iface

/*
 * interface for rpc client
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IClient interface {
	Quit()
	DocSync(tag string, jsonByte []byte) bool
}

