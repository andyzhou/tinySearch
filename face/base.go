package face

import (
	"bytes"
	"github.com/andyzhou/tinysearch/json"
	"github.com/blevesearch/bleve/v2/document"
	"github.com/blevesearch/bleve/v2/search"
	index "github.com/blevesearch/bleve_index_api"
	"io/ioutil"
)

/*
 * face for base
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Base struct {
}

//get sub dirs for data path
func (f *Base) GetSubDirs(dataPath string) ([]string, error) {
	filesInfo, err := ioutil.ReadDir(dataPath)
	if err != nil {
		return nil, err
	}
	if filesInfo == nil || len(filesInfo) <= 0 {
		return nil, nil
	}
	//format result
	result := make([]string, 0)
	for _, v := range filesInfo {
		if v.IsDir() {
			result = append(result, v.Name())
		}
	}
	return result, nil
}

//analyze doc with hit
func (f *Base) AnalyzeDoc(
		doc index.Document,
		hit *search.DocumentMatch,
	) (*json.HitDocJson, error) {
	var (
		jsonByte []byte
		err error
	)

	//init hit doc json
	hitDocJson := json.NewHitDocJson()

	if doc != nil {
		//init one doc object
		jsonObj := json.NewBaseJson()
		genMap := f.FormatDoc(doc)
		if genMap == nil {
			return nil, nil
		}

		//get json byte
		jsonByte, err = jsonObj.EncodeSimple(genMap)
		if err != nil {
			return nil, err
		}
	}

	//set doc json fields
	if hit != nil {
		hitDocJson.Id = hit.ID
		hitDocJson.Score = hit.Score
	}else{
		hitDocJson.Id = doc.ID()
	}
	if jsonByte != nil {
		hitDocJson.OrgJson = jsonByte
	}

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
		doc index.Document,
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
	doc.VisitFields(func(field index.Field) {
		fieldName = field.Name()
		switch field.(type) {
		case *document.TextField:
			{
				v, ok := field.(*document.TextField)
				if ok {
					genMap[fieldName] = string(v.Value())
				}
			}
			break
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
			break
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
			break
		case *document.DateTimeField:
			{
				v, ok := field.(*document.DateTimeField)
				if ok {
					dateValue, _, err := v.DateTime()
					if err == nil {
						genMap[fieldName] = dateValue.Unix()
					}
				}
			}
			break
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
			break
		}
	})
	return genMap
}