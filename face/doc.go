package face

import (
	"errors"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/json"
)

/*
 * face for doc
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Doc struct {
	hookForAddDoc func(jsonByte []byte) error
	Base
}

//construct
func NewDoc() *Doc {
	//self init
	this := &Doc{}
	return this
}

//get doc count
func (f *Doc) GetCount(
		index iface.IIndex,
	) (int64, error) {
	var (
		count int64
	)

	//basic check
	if index == nil {
		return count, errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return count, errors.New("cant' get index")
	}

	//get doc count
	v, err := indexer.DocCount()
	if err != nil {
		return count, err
	}
	return int64(v), nil
}

//remove batch docs
func (f *Doc) RemoveDocs(
		index iface.IIndex,
		docIds ...string,
	) error {
	var (
		err error
	)

	//basic check
	if index == nil || docIds == nil || len(docIds) <= 0 {
		return errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return errors.New("cant' get index")
	}

	//remove one by one
	for _, docId := range docIds {
		err = indexer.Delete(docId)
	}
	return err
}

//remove doc
func (f *Doc) RemoveDoc(
		index iface.IIndex,
		docId string,
	) error {
	//basic check
	if index == nil || docId == "" {
		return errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return errors.New("cant' get index")
	}

	//remove doc
	err := indexer.Delete(docId)
	if err != nil {
		return err
	}
	return nil
}

//get batch docs by id
func (f *Doc) GetDocs(
		index iface.IIndex,
		docIds ...string,
	) (map[string]*json.HitDocJson, error) {
	//basic check
	if index == nil || docIds == nil {
		return nil, errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return nil, errors.New("cant' get index")
	}

	//get batch doc by ids
	result := make(map[string]*json.HitDocJson)
	for _, docId := range docIds {
		doc, err := indexer.Document(docId)
		if err != nil || doc == nil {
			continue
		}
		hitJson, err := f.AnalyzeDoc(doc, nil)
		if err != nil || hitJson == nil {
			continue
		}
		result[docId] = hitJson
	}
	return result, nil
}

//get one doc by id
func (f *Doc) GetDoc(
		index iface.IIndex,
		docId string,
	) (*json.HitDocJson, error) {
	//basic check
	if index == nil || docId == "" {
		return nil, errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return nil, errors.New("cant' get index")
	}

	//get and check doc
	doc, err := indexer.Document(docId)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	//analyze doc
	return f.AnalyzeDoc(doc, nil)
}

//add new doc
func (f *Doc) AddDoc(
		index iface.IIndex,
		docId string,
		jsonObj interface{},
	) error {
	var (
		err error
	)
	//basic check
	if index == nil || docId == "" || jsonObj == nil {
		return errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return errors.New("cant' get index")
	}

	//add or update doc
	err = indexer.Index(docId, jsonObj)
	return err
}

//get hook for add doc
func (f *Doc) GetHoodForAddDoc() func(jsonByte []byte) error{
	return f.hookForAddDoc
}

//set hook for add doc
//used for opt obj from outside
func (f *Doc) SetHookForAddDoc(
		hook func(jsonByte []byte) error,
	) error {
	//check
	if hook == nil {
		return errors.New("invalid parameter")
	}
	f.hookForAddDoc = hook
	return nil
}