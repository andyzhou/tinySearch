package json

/*
 * json for doc
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//doc json
//type DocJson struct {
//	Id string `json:"id"`
//	JsonObj interface{} `json:"jsonObj"` //original json object
//	BaseJson
//}

//hit doc json
type HitDocJson struct {
	Id         string            `json:"id"`
	HighLights map[string]string `json:"highLights"`
	OrgJson    []byte            `json:"orgJson"`
	Score      float64           `json:"score"`
	BaseJson
}

///////////
//construct
///////////

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
