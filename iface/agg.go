package iface

import "github.com/andyzhou/tinySearch/json"

/*
 * interface for agg
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IAgg interface {
	GetAggList(index IIndex, opt *json.QueryOptJson) *json.AggregatesJson
}
