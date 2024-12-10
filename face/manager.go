package face

import (
	"errors"
	"github.com/andyzhou/tinysearch/iface"
	"sync"
)

/*
 * inter manager for rpc service
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - sync doc for new or remove
 */

//face info
type Manager struct {
	//inter data
	dataPath string
	dictFile string
	indexes  *sync.Map //tag -> IIndex
	//sub face
	doc     iface.IDoc
	query   iface.IQuery
	agg     iface.IAgg
	suggest iface.ISuggest
	Base
}

//construct
func NewManager(
	dataPath string,
	dictFile ...string) *Manager{
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
	}
	//sub face init
	this.suggest = NewSuggest(this)
	this.query = NewQuery(this.suggest)
	this.agg = NewAgg(this.query)
	return this
}

//quit
func (f *Manager) Quit() {
	f.suggest.Quit()
}

//get dict file
func (f *Manager) GetDictFile() string {
	return f.dictFile
}

//set index data path
func (f *Manager) SetDataPath(path string) {
	f.dataPath = path
}

//set dict file path
func (f *Manager) SetDictFile(filePath string) {
	if filePath == "" {
		return
	}
	f.dictFile = filePath
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
func (f *Manager) RemoveIndex(tag string) error {
	//basic check
	if tag == "" || f.indexes == nil {
		return errors.New("invalid tag or index is nil")
	}
	//remove index
	f.indexes.Delete(tag)
	return nil
}

//get search index
func (f *Manager) GetIndex(tag string) iface.IIndex {
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
func (f *Manager) AddIndex(tag string) error {
	var (
		err error
	)

	//basic check
	if tag == "" {
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