package json

import "encoding/json"

/*
 * json for agg
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//agg field
type AggField struct {
	Field         string      `json:"field"`
	Size          int         `json:"size"`
	IsNumeric     bool        `json:"isNumeric"`
	NumericRanges []*RangeVal `json:"numericRanges"`
}

//aggregate json
type AggregateJson struct {
	Name  string      `json:"name"`
	Min   json.Number `json:"min,string,omitempty"`
	Max   json.Number `json:"max,string,omitempty"`
	Count int         `json:"count"`
	BaseJson
}

//one kind agg record
type AggregatesJson struct {
	Field   string                      `json:"field"`
	MapList map[string][]*AggregateJson `json:"list"`
	BaseJson
}

///////////////////////////
//construct for AggregatesJson
//////////////////////////

func NewAggregatesJson() *AggregatesJson {
	this := &AggregatesJson{
		MapList: make(map[string][]*AggregateJson, 0),
	}
	return this
}

//add obj
func (j *AggregatesJson) AddObj(aggName string, obj *AggregateJson) bool {
	if aggName == "" || obj == nil {
		return false
	}
	v, ok := j.MapList[aggName]
	if !ok || v == nil {
		v = make([]*AggregateJson, 0)
	}
	v = append(v, obj)
	j.MapList[aggName] = v
	return true
}

//encode json data
func (j *AggregatesJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *AggregatesJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}

///////////////////////////
//construct for AggregateJson
//////////////////////////

func NewAggregateJson() *AggregateJson {
	this := &AggregateJson{
	}
	return this
}

//encode json data
func (j *AggregateJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *AggregateJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}
