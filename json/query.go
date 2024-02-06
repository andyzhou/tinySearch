package json

import (
	"time"
)

/*
 * query opt json
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//range value
type RangeVal struct {
	From float64 `json:"from"`
	To   float64 `json:"to"`
}

//filter field
type FilterField struct {
	Kind        int         `json:"kind"`
	Field       string      `json:"field"`
	Val         interface{} `json:"val"`
	DocIds      []string    `json:"docIds"`      //used for batch doc ids match
	MinVal      string      `json:"minVal"`      //for term range
	MaxVal      string      `json:"maxVal"`      //for term range
	MinFloatVal float64     `json:"minFloatVal"` //for numeric range
	MaxFloatVal float64     `json:"maxFloatVal"` //for numeric range
	StartTime   time.Time   `json:"startTime"`   //for date range
	EndTime     time.Time   `json:"endTime"`     //for date range
	Terms       []string    `json:"terms"`       //for terms query
	IsMust      bool        `json:"isMust"`
	IsExclude   bool        `json:"isExclude"`
}

//sort field
type SortField struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"` //true:desc false:asc
}

//term query para
type TermQueryPara struct {
	Field string `json:"field"`
	Val   string `json:"val"`
}

//json info
type QueryOptJson struct {
	QueryKind  int            `json:"queryKind"`
	TermPara   TermQueryPara  `json:"termPara"`
	Tag        string         `json:"tag"`
	SuggestTag string         `json:"suggestTag"`
	Key        string         `json:"key"`
	Fields     []string       `json:"fields"`
	Filters    []*FilterField `json:"filters"`   //sub filters
	AggFields  []*AggField    `json:"aggFields"` //only for agg
	Sort       []*SortField   `json:"sort"`
	HighLight  bool           `json:"highLight"`
	Offset     int            `json:"offset"` //first priority
	Size       int            `json:"size"`
	Page       int            `json:"page"` //second priority
	PageSize   int            `json:"pageSize"`
	AggSize    int            `json:"aggSize"`
	NeedDocs   bool           `json:"needDocs"`
	Lon        float64        `json:"lon"` //geo of lon
	Lat        float64        `json:"lat"` //geo of lat
	Distance   string         `json:"distance"` //like '1km'
	BaseJson
}


///////////////////////////
//construct for FilterField
//////////////////////////

func NewFilterField() *FilterField {
	this := &FilterField{
		DocIds:make([]string, 0),
		Terms: []string{},
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
		AggFields:make([]*AggField, 0),
		Sort:make([]*SortField, 0),
	}
	return this
}

//gen one agg field
func (j *QueryOptJson) GenAggField() *AggField{
	return &AggField{
		NumericRanges: []*RangeVal{},
	}
}

//add agg field
func (j *QueryOptJson) AddAggField(agg ... *AggField) bool {
	if agg == nil || len(agg) <= 0 {
		return false
	}
	j.AggFields = append(j.AggFields, agg...)
	return true
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