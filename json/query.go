package json

import (
	"encoding/json"
	"time"
)

/*
 * query opt json
 */

//range value
type RangeVal struct {
	From float64 `json:"from"`
	To float64 `json:"to"`
}

//filter field
type FilterField struct {
	Kind int `json:"kind"`
	Field string `json:"field"`
	Val interface{} `json:"val"`
	DocIds []string `json:"docIds"` //used for batch doc ids match
	MinVal string `json:"minVal"` //for term range
	MaxVal string `json:"maxVal"` //for term range
	MinFloatVal json.Number `json:"minFloatVal,string,omitempty"` //for numeric range
	MaxFloatVal json.Number `json:"maxFloatVal,string,omitempty"` //for numeric range
	StartTime time.Time `json:"startTime"` //for date range
	EndTime time.Time `json:"endTime"` //for date range
	IsMust bool `json:"isMust"`
	IsExclude bool `json:"isExclude"`
}

//sort field
type SortField struct {
	Field string `json:"field"`
	Desc bool `json:"desc"` //true:desc false:asc
}

//term query para
type TermQueryPara struct {
	Field string `json:"field"`
	Val string `json:"val"`
}

//json info
type QueryOptJson struct {
	QueryKind int `json:"queryKind"`
	TermPara TermQueryPara `json:"termPara"`
	Tag string `json:"tag"`
	SuggestTag string `json:"suggestTag"`
	Key string `json:"key"`
	Fields []string `json:"fields"`
	Filters []*FilterField `json:"filters"`
	AggField *AggField `json:"aggField"` //only for agg
	Sort []*SortField `json:"sort"`
	HighLight bool `json:"highLight"`
	Page int `json:"page"`
	PageSize int `json:"pageSize"`
	AggSize int `json:"aggSize"`
	BaseJson
}


///////////////////////////
//construct for FilterField
//////////////////////////

func NewFilterField() *FilterField {
	this := &FilterField{
		DocIds:make([]string, 0),
	}
	return this
}

///////////////////////////
//construct for QueryOptJson
//////////////////////////

func NewQueryOptJson() *QueryOptJson {
	this := &QueryOptJson{
		Fields: make([]string, 0),
		Filters: make([]*FilterField, 0),
		AggField: &AggField{
			NumericRanges: make([]*RangeVal, 0),
		},
		Sort:make([]*SortField, 0),
	}
	return this
}

//add field
func (j *QueryOptJson) AddField(field ... string) bool {
	if field == nil {
		return false
	}
	j.Fields = append(j.Fields, field...)
	return true
}

//add filter
func (j *QueryOptJson) AddFilter(obj ... *FilterField) bool {
	if obj == nil {
		return false
	}
	j.Filters = append(j.Filters, obj...)
	return true
}

//encode json data
func (j *QueryOptJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *QueryOptJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}