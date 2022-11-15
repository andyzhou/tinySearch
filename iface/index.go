package iface

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

/*
 * interface for index
 */

type IIndex interface {
	RemoveIndex() error
	GetIndex() bleve.Index
	CreateIndex() error
	CreateChineseMap(dictPath string) (*mapping.IndexMappingImpl, error)
	SetDictPath(dict string) bool
}
