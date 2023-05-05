package iface

import "github.com/andyzhou/tinysearch/json"

/*
 * interface for suggest
 */

type ISuggest interface {
	Quit()
	GetSuggest(opt *json.SuggestOptJson) (*json.SuggestsJson, error)
	AddSuggest(tag string, doc *json.SuggestJson) error
	RegisterSuggest(tags ...string) error
}
