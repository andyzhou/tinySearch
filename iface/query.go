package iface

import "github.com/andyzhou/tinySearch/json"

/*
 * interface for query
 */

type IQuery interface {
	Query(index IIndex, json *json.QueryOptJson) (*json.SearchResultJson, error)
}