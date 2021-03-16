package face

import (
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve_index_api"
)

/*
 * face for base
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Base struct {
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
	})
	return genMap
}