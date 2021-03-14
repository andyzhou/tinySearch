package iface

import "github.com/andyzhou/tinySearch/json"

/*
 * interface for doc
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IDoc interface {
	GetCount(index IIndex) (int64, error)
	RemoveDoc(index IIndex, docId string) error
	AddDoc(index IIndex, obj *json.DocJson) error
}