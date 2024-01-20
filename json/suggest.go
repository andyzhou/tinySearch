package json

/*
 * json for suggest
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//suggest opt json
type SuggestOptJson struct {
	QueryKind int    `json:"queryKind"`
	IndexTag  string `json:"indexTag"`
	Key       string `json:"key"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	BaseJson
}

//suggest doc json
type SuggestJson struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
	BaseJson
}

type SuggestsJson struct {
	Total int64          `json:"total"`
	List  []*SuggestJson `json:"list"`
	BaseJson
}


///////////////////////////
//construct for SuggestsJson
//////////////////////////

func NewSuggestsJson() *SuggestsJson {
	this := &SuggestsJson{
		List: make([]*SuggestJson, 0),
	}
	return this
}

 //add obj
func (j *SuggestsJson) AddObj(obj *SuggestJson) bool {
	if obj == nil {
		return false
	}
	j.List = append(j.List, obj)
	return true
}

//encode json data
func (j *SuggestsJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *SuggestsJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}

///////////////////////////
//construct for SuggestJson
//////////////////////////

func NewSuggestJson() *SuggestJson {
	this := &SuggestJson{
	}
	return this
}

//encode json data
func (j *SuggestJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *SuggestJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}


///////////////////////////
//construct for SuggestOptJson
//////////////////////////

func NewSuggestOptJson() *SuggestOptJson {
	this := &SuggestOptJson{
	}
	return this
}

//encode json data
func (j *SuggestOptJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *SuggestOptJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}