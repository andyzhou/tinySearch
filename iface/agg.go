package iface

import "github.com/andyzhou/tinysearch/json"

/*
 * interface for agg
 */

type IAgg interface {
	GetAggList(index IIndex, opt *json.QueryOptJson) (*json.AggregatesJson, error)
}
