package face

import (
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
			) int64 {
	var (
		count int64
	)

	//basic check
	if index == nil {
		return count
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return count
	}

	//get doc count
	v, err := (*indexer).DocCount()
	if err != nil {
		log.Println("Doc::GetCount failed, err:", err.Error())
		return count
	}
	return int64(v)
}

//remove doc
func (f *Doc) RemoveDoc(
				index iface.IIndex,
				docId string,
			) bool {
	//basic check
	if index == nil || docId == "" {
		return false
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return false
	}

	//remove doc
	err := (*indexer).Delete(docId)
	if err != nil {
		log.Println("Doc::RemoveDoc failed, err:", err.Error())
		return false
	}
	return true
}

//add new doc
func (f *Doc) AddDoc(
				index iface.IIndex,
				obj *json.DocJson,
			) bool {
	//basic check
	if index == nil || obj == nil {
		return false
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return false
	}

	//add or update doc
	err := (*indexer).Index(obj.Id, obj.JsonObj)
	if err != nil {
		log.Println("Doc::AddDoc failed, err:", err.Error())
		return false
	}

	return true
}