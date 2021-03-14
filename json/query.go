package json

import "time"

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
	MinVal float64 `json:"minVal"` //for numeric range
	MaxVal float64 `json:"maxVal"` //for numeric range
	StartTime time.Time `json:"startTime"` //for date range
	EndTime time.Time `json:"endTime"` //for date range
}

//sort field
type SortField struct {
	Field string `json:"field"`
	Ascending bool `json:"ascending"` //true:asc false:dsc
}

//json info
type QueryOptJson struct {
	Tag string `json:"tag"`
	Key string `json:"key"`
	Fields []string `json:"fields"`
	Filters []*FilterField `json:"filters"`
	AggField *AggField `json:"aggField"` //only for agg
	Sort *SortField `json:"sort"`
	HighLight bool `json:"highLight"`
	Page int `json:"page"`
	PageSize int `json:"pageSize"`
	AggSize int `json:"aggSize"`
	BaseJson
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
func (j *QueryOptJson) Encode() []byte {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *QueryOptJson) Decode(data []byte) bool {
	return j.BaseJson.Decode(data, j)
}