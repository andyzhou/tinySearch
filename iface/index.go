package iface

import "github.com/blevesearch/bleve/v2"

/*
 * interface for index
 */

type IIndex interface {
	RemoveIndex() bool
	GetIndex() bleve.Index
	CreateIndex() error
}
