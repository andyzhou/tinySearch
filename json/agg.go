package json

/*
 * json for agg
 */

//agg field
type AggField struct {
	Field string `json:"field"`
	Size int `json:"size"`
	IsNumeric bool `json:"isNumeric"`
	NumericRanges []*RangeVal `json:"numericRanges"`
}

//aggregate json
type AggregateJson struct {
	Name string `json:"name"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Count int `json:"count"`
	BaseJson
}

//one kind agg record
type AggregatesJson struct {
	Field string `json:"field"`
	List []*AggregateJson `json:"list"`
	BaseJson
}

///////////////////////////
//construct for AggregatesJson
//////////////////////////

func NewAggregatesJson() *AggregatesJson {
	this := &AggregatesJson{
		List: make([]*AggregateJson, 0),
	}
	return this
}

//add obj
func (j *AggregatesJson) AddObj(obj *AggregateJson) bool {
	if obj == nil {
		return false
	}
	j.List = append(j.List, obj)
	return true
}

//encode json data
func (j *AggregatesJson) Encode() []byte {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *AggregatesJson) Decode(data []byte) bool {
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
func (j *AggregateJson) Encode() []byte {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *AggregateJson) Decode(data []byte) bool {
	return j.BaseJson.Decode(data, j)
}
