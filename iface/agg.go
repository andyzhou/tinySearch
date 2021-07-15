package iface

import "github.com/andyzhou/tinySearch/json"

/*
 * interface for agg
 */

type IAgg interface {
	GetAggList(index IIndex, opt *json.QueryOptJson) (*json.AggregatesJson, error)
}
