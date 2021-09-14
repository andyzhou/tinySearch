package iface

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

/*
 * interface for index
 */

type IIndex interface {
	RemoveIndex() bool
	GetIndex() bleve.Index
	CreateIndex(indexMap ...*mapping.IndexMappingImpl) error
	CreateChineseMap(dictPath string) (*mapping.IndexMappingImpl, error)
}
