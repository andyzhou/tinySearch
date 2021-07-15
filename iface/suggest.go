package iface

import "github.com/andyzhou/tinySearch/json"

/*
 * interface for suggest
 */

type ISuggest interface {
	Quit()
	GetSuggest(opt *json.SuggestOptJson) (*json.SuggestsJson, error)
	AddSuggest(doc *json.SuggestJson) bool
}
