package face

import (
	"bytes"
	"errors"
	"github.com/andyzhou/tinySearch/json"
	"github.com/andyzhou/tinycells/tc"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/search"
)

/*
 * face for base
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Base struct {
}

//analyze doc with hit
func (f *Base) AnalyzeDoc(
				doc *document.Document,
				hit *search.DocumentMatch,
			) (*json.HitDocJson, error) {
	//basic check
	if doc == nil {
		return nil, errors.New("invalid parameter")
	}

	//init one doc object
	jsonObj := tc.NewBaseJson()
	genMap := f.FormatDoc(doc)
	if genMap == nil {
		return nil, nil
	}

	//get json byte
	jsonByte := jsonObj.EncodeSimple(genMap)

	//init hit doc json
	hitDocJson := json.NewHitDocJson()

	//set doc json fields
	hitDocJson.Id = doc.ID
	hitDocJson.OrgJson = jsonByte

	//check high light
	if hit != nil && hit.Fragments != nil {
		buffer := bytes.NewBuffer(nil)
		for k, v := range hit.Fragments {
			buffer.Reset()
			for _, v1 := range v {
				buffer.WriteString(v1)
			}
			hitDocJson.AddHighLight(k, buffer.String())
		}
	}
	return hitDocJson, nil
}

//format one doc
func (f *Base) FormatDoc(
				doc *document.Document,
			) map[string]interface{} {
	var (
		fieldName string
	)

	//basic check
	if doc == nil {
		return nil
	}

	//format result
	genMap := make(map[string]interface{})

	//analyze fields
	for _, field := range doc.Fields {
		fieldName = field.Name()
		switch field.(type) {
		case *document.TextField:
			{
				genMap[fieldName] = string(field.Value())
			}
		case *document.NumericField:
			{
				v, ok := field.(*document.NumericField)
				if ok {
					numericValue, err := v.Number()
					if err == nil {
						genMap[fieldName] = numericValue
					}
				}
			}
		case *document.BooleanField:
			{
				v, ok := field.(*document.BooleanField)
				if ok {
					boolValue, err := v.Boolean()
					if err == nil {
						genMap[fieldName] = boolValue
					}
				}
			}
		case *document.DateTimeField:
			{
				v, ok := field.(*document.DateTimeField)
				if ok {
					dateValue, err := v.DateTime()
					if err == nil {
						genMap[fieldName] = dateValue.Unix()
					}
				}
			}
		case *document.GeoPointField:
			{
				v, ok := field.(*document.GeoPointField)
				if ok {
					latVal, _ := v.Lat()
					lonVal, _ := v.Lon()
					genMap[fieldName] = []interface{}{
						latVal,
						lonVal,
					}
				}
			}
		}
	}
	return genMap
}