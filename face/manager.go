package face

import (
	"errors"
	"github.com/andyzhou/tinySearch/iface"
	"sync"
)

/*
 * inter manager for rpc service
 * - sync doc for new or remove
 */

//face info
type Manager struct {
	dataPath string
	dictFile string
	indexes *sync.Map
	//sub face
	doc iface.IDoc
	query iface.IQuery
	agg iface.IAgg
	suggest iface.ISuggest
}

//construct
func NewManager(dataPath string, dictFile ...string) *Manager{
	var (
		dictFilePath string
	)
	//get dict file
	if dictFile != nil && len(dictFile) > 0 {
		dictFilePath = dictFile[0]
	}
	//self init
	this := &Manager{
		dataPath:dataPath,
		dictFile: dictFilePath,
		indexes:new(sync.Map),
		doc:NewDoc(),
		suggest:NewSuggest(dataPath, dictFilePath),
	}
	this.query = NewQuery(this.suggest)
	this.agg = NewAgg(this.query)
	return this
}

//quit
func (f *Manager) Quit() {
}

//get sub face
func (f *Manager) GetDoc() iface.IDoc {
	return f.doc
}

func (f *Manager) GetQuery() iface.IQuery {
	return f.query
}

func (f *Manager) GetAgg() iface.IAgg {
	return f.agg
}

func (f *Manager) GetSuggest() iface.ISuggest {
	return f.suggest
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
	index, ok := v.(iface.IIndex)
	if !ok {
		return nil
	}
	return index
}

//add search index
func (f *Manager) AddIndex(
					tag string,
				) error {
	var (
		err error
	)
	//basic check
	if tag == "" || f.indexes == nil {
		return errors.New("invalid parameter")
	}

	//check record
	_, ok := f.indexes.Load(tag)
	if ok {
		return nil
	}

	//init new index
	index := NewIndex(f.dataPath, tag, f.dictFile)
	err = index.CreateIndex()
	if err != nil {
		return err
	}

	//sync into map
	f.indexes.Store(tag, index)
	return nil
}