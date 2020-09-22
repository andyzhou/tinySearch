package face

import (
	"github.com/andyzhou/tinySearch/iface"
	"sync"
)

/*
 * face for inter manager
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Manager struct {
	indexes *sync.Map
	clients *sync.Map
}

//construct
func NewManager() *Manager{
	//self init
	this := &Manager{
		indexes:new(sync.Map),
		clients:new(sync.Map),
	}
	return this
}

//quit
func (f *Manager) Quit() {
	if f.clients == nil {
		return
	}
	sf := func(_, v interface{}) bool {
		client, ok := v.(*Client)
		if !ok {
			return false
		}
		client.Quit()
		return true
	}
	f.clients.Range(sf)
}

//doc sync to all clients
func (f *Manager) DocSync(
					tag string,
					jsonByte []byte,
				) bool {
	//basic check
	if tag == "" || jsonByte == nil {
		return false
	}

	if f.clients == nil {
		return false
	}

	//do doc sync on all clients
	sf := func(k, v interface{}) bool {
		client, ok := v.(*Client)
		if !ok {
			return false
		}
		client.DocSync(tag, jsonByte)
		return true
	}
	f.clients.Range(sf)

	return true
}

//remove client node
func (f *Manager) RemoveNode(
					addr string,
				) bool {
	//basic check
	if addr == "" || f.clients == nil {
		return false
	}

	//remove
	f.clients.Delete(addr)

	return true
}

//add client node
func (f *Manager) AddNode(
					addr string,
				) bool {
	//basic check
	if addr == "" || f.clients == nil {
		return false
	}

	//check record
	_, ok := f.clients.Load(addr)
	if ok {
		return false
	}

	//init new client
	client := NewClient(addr)

	//sync into map
	f.clients.Store(addr, client)

	return true
}

////////////////
//api for index
////////////////

//remove index
func (f *Manager) RemoveIndex(
					tag string,
				) bool {
	//basic check
	if tag == "" || f.indexes == nil {
		return false
	}

	//remove index
	f.indexes.Delete(tag)

	return true
}

//get search index
func (f *Manager) GetIndex(
					tag string,
				) iface.IIndex {
	//basic check
	if tag == "" || f.indexes == nil {
		return nil
	}

	//load record
	v, ok := f.indexes.Load(tag)
	if !ok {
		return nil
	}
	index, ok := v.(*Index)
	if !ok {
		return nil
	}

	return index
}

//add search index
func (f *Manager) AddIndex(
					dir string,
					tag string,
				) bool {
	//basic check
	if dir == "" || tag == "" || f.indexes == nil {
		return false
	}

	//check record
	_, ok := f.indexes.Load(tag)
	if ok {
		return false
	}

	//init new index
	index := NewIndex(dir, tag)

	//sync into map
	f.indexes.Store(tag, index)

	return true
}