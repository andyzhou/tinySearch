package iface

/*
 * interface for inter manager
 */

type IManager interface {
	Quit()

	//for index
	RemoveIndex(tag string) bool
	GetIndex(tag string) IIndex
	AddIndex(tag string, useChineseTokenizer ...bool) error
	SetDictPath(dict string) bool

	//get sub face
	GetDoc() IDoc
	GetQuery() IQuery
	GetAgg() IAgg
	GetSuggest() ISuggest
}