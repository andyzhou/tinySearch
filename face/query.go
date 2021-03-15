package face

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/andyzhou/tinySearch/define"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	"github.com/andyzhou/tinycells/tc"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"log"
)

/*
 * face for query
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Query struct {
	Base
}

//construct
func NewQuery() *Query {
	//self init
	this := &Query{
	}
	return this
}

//query doc
func (f *Query) Query(
					index iface.IIndex,
					opt *json.QueryOptJson,
				) (*json.SearchResultJson, error) {
	var (
		tempStr string
		docQuery *query.MatchQuery
		searchRequest *bleve.SearchRequest
	)

	//basic check
	if index == nil || opt == nil {
		return nil, errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return nil, errors.New("can't get indexer")
	}

	//init query
	if opt.Key != "" {
		docQuery = bleve.NewMatchQuery(opt.Key)
	}else{
		docQuery = bleve.NewMatchQuery("")
	}

	//set query fields
	if opt.Fields != nil && len(opt.Fields) > 0 {
		for _, field := range opt.Fields {
			//set query field
			docQuery.SetField(field)
		}
	}

	//set filter fields
	if opt.Filters != nil && len(opt.Filters) > 0 {
		//init bool query
		boolQuery := bleve.NewBooleanQuery()

		//add filter field and value
		for _, filter := range opt.Filters {
			//do sub query by kind
			switch filter.Kind {
			case define.FilterKindMatch:
				{
					//match by condition
					tempStr = fmt.Sprintf("%v", filter.Val)
					pg := bleve.NewTermQuery(tempStr)
					pg.SetField(filter.Field)
					boolQuery.AddMust(pg)
				}
			case define.FilterKindQuery:
				{
					//sub phrase query
					tempStr = fmt.Sprintf("%v", filter.Val)
					pq := bleve.NewPhraseQuery([]string{tempStr}, filter.Field)
					boolQuery.AddMust(pq)
				}
			case define.FilterKindNumericRange:
				{
					//min <= val < max
					pg := bleve.NewNumericRangeQuery(&filter.MinVal, &filter.MaxVal)
					pg.SetField(filter.Field)
					boolQuery.AddMust(pg)
				}
			case define.FilterKindDateRange:
				{
					pg := bleve.NewDateRangeQuery(filter.StartTime, filter.EndTime)
					pg.SetField(filter.Field)
					boolQuery.AddMust(pg)
				}
			}
		}

		//add should query
		boolQuery.AddMust(docQuery)

		//init multi condition search request
		searchRequest = bleve.NewSearchRequest(boolQuery)
	}else{
		//general search request
		searchRequest = bleve.NewSearchRequest(docQuery)
	}

	//set high light
	if opt.HighLight {
		//other search request
		searchRequest.Highlight = bleve.NewHighlight()
	}

	//check page and page size
	if opt.Page <= 0 {
		opt.Page = 1
	}
	if opt.PageSize <= 0 {
		opt.PageSize = define.RecPerPage
	}

	//set others
	searchRequest.From = (opt.Page - 1) * opt.PageSize
	searchRequest.Size = opt.PageSize
	searchRequest.Explain = true

	//begin search
	searchResult, err := (*indexer).Search(searchRequest)
	if err != nil {
		log.Println("Query::Query failed, err:", err.Error())
		return nil, err
	}

	//check result
	if searchResult.Total <= 0 {
		return nil, nil
	}

	//init result
	result := json.NewSearchResultJson()
	result.Total = searchResult.Total

	//format records
	result.Records = f.formatResult(indexer, &searchResult.Hits)

	return result, nil
}

///////////////
//private func
///////////////

//format result
func (f *Query) formatResult(
					index *bleve.Index,
					hits *search.DocumentMatchCollection,
				) []*json.HitDocJson {
	var (
		buffer *bytes.Buffer
	)

	//basic check
	if index == nil || hits == nil {
		return nil
	}

	//format result
	result := make([]*json.HitDocJson, 0)

	//format records
	for _, hit := range *hits {
		//get original doc
		doc, err := (*index).Document(hit.ID)
		if err != nil {
			continue
		}

		//init one doc object
		jsonObj := tc.NewBaseJson()
		genMap := f.FormatDoc(doc)
		if genMap == nil {
			continue
		}

		//get json byte
		jsonByte := jsonObj.EncodeSimple(genMap)

		//init hit doc json
		hitDocJson := json.NewHitDocJson()

		//set doc json fields
		hitDocJson.Id = hit.ID
		hitDocJson.OrgJson = jsonByte

		//check high light
		if hit.Fragments != nil {
			buffer = bytes.NewBuffer(nil)
			for k, v := range hit.Fragments {
				buffer.Reset()
				for _, v1 := range v {
					buffer.WriteString(v1)
				}
				hitDocJson.AddHighLight(k, buffer.String())
			}
		}

		//add into slice
		result = append(result, hitDocJson)
	}

	return result
}
