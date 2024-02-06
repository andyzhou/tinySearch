package json

import "github.com/andyzhou/tinysearch/json"

//testing doc json
type TestDocJson struct {
	Id        int64                  `json:"id,string,omitempty"` //for fix int64 Unmarshal issue
	Title     string                 `json:"title"`
	Cat       string                 `json:"cat"`
	CatPath   string                 `json:"catPath"`
	Price     float64                `json:"price"` //need use match
	Num       int64                  `json:"num"`
	PosterId  string                 `json:"posterId"`
	Prop      map[string]interface{} `json:"prop"`
	Tags      []string               `json:"tags"`
	Introduce string                 `json:"introduce"`
	CreateAt  int64                  `json:"createAt"`
	json.BaseJson
}

///////////////////////////
//construct for TestDocJson
//////////////////////////

func NewTestDocJson() *TestDocJson {
	this := &TestDocJson{
		Prop: make(map[string]interface{}),
		Tags: make([]string, 0),
	}
	return this
}

//encode json data
func (j *TestDocJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *TestDocJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}
