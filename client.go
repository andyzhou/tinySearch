package tinySearch

import (
	"errors"
	"github.com/andyzhou/tinySearch/face"
	"github.com/andyzhou/tinySearch/iface"
)

/*
 * client api
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//query opt kind
const (
	QueryOptKindOfGen = iota
	QueryOptKindOfAgg
	QueryOptKindOfSuggest
)

//face info
type Client struct {
	manager iface.IManager
}

//construct
func NewClient() *Client {
	//self init
	self := &Client{
		manager: face.NewManager(""),
	}
	return self
}

//quit
func (f *Client) Quit() {
	f.manager.Quit()
}

//query doc
func (f *Client) DocQuery(optKind int, tag string, optJson []byte) ([][]byte, int32, error) {
	//get client
	client := f.manager.GetClient()
	if client == nil {
		return nil, 0, errors.New("can't get client")
	}

	//doc query
	return client.DocQuery(optKind, tag, optJson)
}

//remove doc
func (f *Client) DocRemove(tag, docId string) error {
	//remove doc from relate nodes pass manager
	err := f.manager.DocRemove(tag, docId)
	return err
}

//sync doc
func (f *Client) DocSync(tag, docId string, docJson []byte) error {
	//sync doc to relate nodes pass manager
	err := f.manager.DocSync(tag, docId, docJson)
	return err
}

//add search service nodes
func (f *Client) AddNodes(nodes ... string) bool {
	//check
	if nodes == nil || len(nodes) <= 0 {
		return false
	}

	//add into manager
	bRet := f.manager.AddNode(nodes...)
	return bRet
}