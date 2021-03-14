package iface

import "github.com/blevesearch/bleve"

/*
 * interface for index
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IIndex interface {
	RemoveIndex() bool
	GetIndex() *bleve.Index
	CreateIndex() error
}
