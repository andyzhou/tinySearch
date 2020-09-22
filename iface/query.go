package iface

import "github.com/andyzhou/tinySearch/json"

/*
 * interface for query
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IQuery interface {
	Query(index IIndex, json *json.QueryOptJson) *json.SearchResultJson
}