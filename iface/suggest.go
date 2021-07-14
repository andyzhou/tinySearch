package iface

import "github.com/andyzhou/tinySearch/json"

/*
 * interface for suggest
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type ISuggest interface {
	Quit()
	GetSuggest(opt *json.SuggestOptJson) *json.SuggestsJson
	AddSuggest(doc *json.SuggestJson) bool
}
