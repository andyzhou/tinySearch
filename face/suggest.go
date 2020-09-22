package face

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/andyzhou/tinySearch/define"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	"github.com/blevesearch/bleve"
	"log"
)

/*
 * face for suggest
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//suggest record field
const (
	SuggestFieldKind = "kind"
	SuggestFieldKey = "key"
	SuggestFieldCount = "count"
)

//face info
type Suggest struct {
	Base
}

//construct
func NewSuggest() *Suggest {
	//self init
	this := &Suggest{
	}
	return this
}

//get suggest
func (f *Suggest) GetSuggest(
					index iface.IIndex,
					opt *json.SuggestOptJson,
				) *json.SuggestsJson {
	//basic check
	if index == nil || opt == nil {
		return nil
	}

	//get index
	indexer := index.GetIndex()
	if indexer == nil {
		return nil
	}

	//init query
	docQuery := bleve.NewMatchQuery(opt.Key)

	//set query field
	docQuery.SetField("key")

	//set filter field
	//init bool query
	boolQuery := bleve.NewBooleanQuery()

	//add must query
	boolQuery.AddMust(docQuery)

	//add filter field and value
	kindFloat := float64(opt.Kind)
	kindFloatEnd := kindFloat + 1
	pq := bleve.NewNumericRangeQuery(&kindFloat, &kindFloatEnd)
	pq.SetField("kind")
	boolQuery.AddMust(pq)

	//init multi condition search request
	searchRequest := bleve.NewSearchRequest(boolQuery)

	//set others
	searchRequest.From = 0
	searchRequest.Size = define.RecPerPage
	searchRequest.Explain = true

	//begin search
	searchResult, err := (*indexer).Search(searchRequest)
	if err != nil {
		log.Println("Suggest::GetSuggest failed, err:", err.Error())
		return nil
	}

	//check hits
	if searchResult.Hits == nil ||
		searchResult.Hits.Len() <= 0 {
		return nil
	}

	//init result
	result := json.NewSuggestsJson()

	//format records
	for _, hit := range searchResult.Hits {
		//get original doc by id
		doc, err := (*indexer).Document(hit.ID)
		if err != nil {
			continue
		}

		//init doc json
		suggestJson := json.NewSuggestJson()

		//format fields
		genMap := f.FormatDoc(doc)
		for k, v := range genMap {
			switch k {
			case SuggestFieldKind:
				{
					v1, ok := v.(float64)
					if ok {
						suggestJson.Kind = int(v1)
					}
				}
			case SuggestFieldKey:
				{
					v1, ok := v.(string)
					if ok {
						suggestJson.Key = v1
					}
				}
			case SuggestFieldCount:
				{
					v1, ok := v.(string)
					if ok {
						suggestJson.Count = v1
					}
				}
			}
		}
		//add into slice
		result.AddObj(suggestJson)
	}

	return result
}


//add new suggest
func (f *Suggest) AddSuggest(
					index iface.IIndex,
					doc *json.SuggestJson,
				) bool {
	//basic check
	if index == nil || doc == nil {
		return false
	}

	//get index
	indexer := index.GetIndex()
	if indexer == nil {
		return false
	}

	//add or update doc
	keyMd5 := f.genMd5(doc.Key)
	err := (*indexer).Index(keyMd5, doc)
	if err != nil {
		log.Println("Suggest::AddSuggest failed, err:", err.Error())
		return false
	}
	return true
}

//////////////
//private func
//////////////

func (f *Suggest) genMd5(
					orgString string,
				) string {
	if len(orgString) <= 0 {
		return ""
	}
	m := md5.New()
	m.Write([]byte(orgString))
	return hex.EncodeToString(m.Sum(nil))
}
