package iface

import (
	"github.com/andyzhou/tinysearch/json"
	"github.com/blevesearch/bleve/v2"
)

/*
 * interface for query
 */

type IQuery interface {
	QueryAll(index IIndex, needDoc ...bool) (*json.SearchResultJson, error)
	Query(index IIndex, json *json.QueryOptJson) (*json.SearchResultJson, error)
	BuildSearchReq(json *json.QueryOptJson) *bleve.SearchRequest
}