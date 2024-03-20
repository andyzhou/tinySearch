package iface

import "github.com/andyzhou/tinysearch/json"

/*
 * interface for doc
 */

type IDoc interface {
	GetCount(index IIndex) (int64, error)
	RemoveDocs(index IIndex, docIds ...string) error
	RemoveDoc(index IIndex, docId string) error
	GetDocs(index IIndex, docIds ...string) (map[string]*json.HitDocJson, error)
	GetDoc(index IIndex, docId string) (*json.HitDocJson, error)
	AddDoc(index IIndex, docId string, jsonObj interface{}) error
	SetHookForAddDoc(hook func(jsonByte []byte) error) error
	GetHoodForAddDoc() func(jsonByte []byte) error
}