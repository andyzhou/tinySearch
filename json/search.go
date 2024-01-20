package json

/*
 * json for search
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//search data json
type SearchJson struct {
	App      string `json:"app"`
	Name     string `json:"name"`
	Count    int    `json:"count"` //doc count
	Status   int    `json:"status"`
	CreateAt int64  `json:"createAt"`
	BaseJson
}

//search result json
type SearchResultJson struct {
	Total   uint64        `json:"total"`
	Records []*HitDocJson `json:"records"`
	BaseJson
}

///////////////////////////
//construct for SearchJson
//////////////////////////

func NewSearchJson() *SearchJson {
	this := &SearchJson{
	}
	return this
}

//encode json data
func (j *SearchJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *SearchJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}

///////////////////////////
//construct for SearchResultJson
//////////////////////////

func NewSearchResultJson() *SearchResultJson {
	this := &SearchResultJson{
		Records:make([]*HitDocJson, 0),
	}
	return this
}

//add doc
func (j *SearchResultJson) AddDoc(obj *HitDocJson) bool {
	if obj == nil {
		return false
	}
	j.Records = append(j.Records, obj)
	return true
}

//encode json data
func (j *SearchResultJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *SearchResultJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}


