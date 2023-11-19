package iface

/*
 * interface for inter manager
 */

type IManager interface {
	Quit()
	GetDictFile() string
	SetDataPath(path string)
	SetDictFile(filePath string)

	//for index
	RemoveIndex(tag string) error
	GetIndex(tag string) IIndex
	AddIndex(tag string) error

	//get sub face
	GetDoc() IDoc
	GetQuery() IQuery
	GetAgg() IAgg
	GetSuggest() ISuggest
}