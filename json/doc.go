package json

/*
 * json for doc
 */

//doc json
//type DocJson struct {
//	Id string `json:"id"`
//	JsonObj interface{} `json:"jsonObj"` //original json object
//	BaseJson
//}

//hit doc json
type HitDocJson struct {
	Id string `json:"id"`
	HighLights map[string]string `json:"highLights"`
	OrgJson []byte `json:"orgJson"`
	BaseJson
}

//testing doc json
type TestDocJson struct {
	Id int64 `json:"id,string,omitempty"` //for fix int64 Unmarshal issue
	Title string `json:"title"`
	Cat string `json:"cat"`
	Price float64 `json:"price"`
	Num int64 `json:"num"`
	Prop map[string]interface{} `json:"prop"`
	Introduce string `json:"introduce"`
	CreateAt int64 `json:"createAt"`
	BaseJson
}

///////////////////////////
//construct for TestDocJson
//////////////////////////

func NewTestDocJson() *TestDocJson {
	this := &TestDocJson{
		Prop: make(map[string]interface{}),
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


///////////////////////////
//construct for HitDocJson
//////////////////////////

func NewHitDocJson() *HitDocJson {
	this := &HitDocJson{
		HighLights: make(map[string]string),
		OrgJson: make([]byte, 0),
	}
	return this
}

func (j *HitDocJson) AddHighLight(field, val string) bool {
	if field == "" || val == "" {
		return false
	}
	j.HighLights[field] = val
	return true
}

//encode json data
func (j *HitDocJson) Encode() ([]byte, error) {
	return j.BaseJson.Encode(j)
}

//decode json data
func (j *HitDocJson) Decode(data []byte) error {
	return j.BaseJson.Decode(data, j)
}

///////////////////////////
//construct for DocJson
//////////////////////////

//func NewDocJson() *DocJson {
//	this := &DocJson{
//	}
//	return this
//}
//
////encode json data
//func (j *DocJson) Encode() []byte {
//	return j.BaseJson.Encode(j)
//}
//
////decode json data
//func (j *DocJson) Decode(data []byte) bool {
//	return j.BaseJson.Decode(data, j)
//}

