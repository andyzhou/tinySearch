package face

import (
	"errors"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	"log"
)

/*
 * face for doc
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Doc struct {
}

//construct
func NewDoc() *Doc {
	//self init
	this := &Doc{
	}
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
	v, err := (*indexer).DocCount()
	if err != nil {
		log.Println("Doc::GetCount failed, err:", err.Error())
		return count, err
	}
	return int64(v), nil
}

//remove batch docs
func (f *Doc) RemoveDocs(
				index iface.IIndex,
				docIds []string,
			) error {
	var (
		err error
	)

	//basic check
	if index == nil || docIds == nil {
		return errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return errors.New("cant' get index")
	}

	//remove one by one
	for _, docId := range docIds {
		err = (*indexer).Delete(docId)
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
	err := (*indexer).Delete(docId)
	if err != nil {
		log.Println("Doc::RemoveDoc failed, err:", err.Error())
		return err
	}
	return nil
}

//add new doc
func (f *Doc) AddDoc(
				index iface.IIndex,
				obj *json.DocJson,
			) error {
	//basic check
	if index == nil || obj == nil {
		return errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return errors.New("cant' get index")
	}

	//add or update doc
	err := (*indexer).Index(obj.Id, obj.JsonObj)
	if err != nil {
		log.Println("Doc::AddDoc failed, err:", err.Error())
		return err
	}

	return nil
}