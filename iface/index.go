package iface

import "github.com/blevesearch/bleve/v2"

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
