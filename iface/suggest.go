package iface

import "github.com/andyzhou/tinySearch/json"

/*
 * interface for suggest
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type ISuggest interface {
	GetSuggest(index IIndex, opt *json.SuggestOptJson) *json.SuggestsJson
	AddSuggest(index IIndex, doc *json.SuggestJson) bool
}
