package face

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/json"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
)

/*
 * face for query
 */

//face info
type Query struct {
	suggester iface.ISuggest //refer of parent
	Base
}

//construct
func NewQuery(suggester iface.ISuggest) *Query {
	//self init
	this := &Query{
		suggester:suggester,
	}
	return this
}

//query all doc
func (f *Query) QueryAll(
			index iface.IIndex,
		) (*json.SearchResultJson, error) {
	//basic check
	if index == nil {
		return nil, errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return nil, errors.New("can't get indexer")
	}

	//init search request
	matchAll := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(matchAll)
	searchRequest.Explain = true

	//begin search
	searchResult, err := indexer.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	//check result
	if searchResult.Total <= 0 {
		result := &json.SearchResultJson{
			Total: 0,
			Records: nil,
		}
		return result, nil
	}

	//init result
	result := json.NewSearchResultJson()
	result.Total = searchResult.Total

	//format records
	result.Records = f.formatResult(indexer, &searchResult.Hits)
	return result, nil
}

//query doc
func (f *Query) Query(
					index iface.IIndex,
					opt *json.QueryOptJson,
				) (*json.SearchResultJson, error) {
	//basic check
	if index == nil || opt == nil {
		return nil, errors.New("invalid parameter")
	}

	//get indexer
	indexer := index.GetIndex()
	if indexer == nil {
		return nil, errors.New("can't get indexer")
	}

	//build search request
	searchRequest := f.BuildSearchReq(opt)

	//set high light
	if opt.HighLight {
		//other search request
		searchRequest.Highlight = bleve.NewHighlight()
	}

	//sort by
	if opt.Sort != nil {
		customSort := make([]search.SearchSort, 0)
		for _, sort := range opt.Sort {
			cs := search.SortField{
				Field: sort.Field,
				Desc: sort.Desc,
			}
			customSort = append(customSort, &cs)
		}
		searchRequest.SortByCustom(customSort)
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
	searchResult, err := indexer.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	//check result
	if searchResult.Total <= 0 {
		result := &json.SearchResultJson{
			Total: 0,
			Records: nil,
		}
		return result, nil
	}

	//sync into suggester
	if f.suggester != nil && opt.Key != "" {
		if searchResult.Total > 0 && opt.SuggestTag != "" {
			suggestJson := json.NewSuggestJson()
			suggestJson.Key = opt.Key
			suggestJson.Count = int64(searchResult.Total)
			f.suggester.AddSuggest(opt.SuggestTag, suggestJson)
		}
	}

	//init result
	result := json.NewSearchResultJson()
	result.Total = searchResult.Total

	//format records
	result.Records = f.formatResult(indexer, &searchResult.Hits)

	return result, nil
}

//build query object
func (f *Query) BuildSearchReq(opt *json.QueryOptJson) *bleve.SearchRequest {
	var (
		docQuery query.Query
		searchRequest *bleve.SearchRequest
	)

	//setup search kind
	switch opt.QueryKind {
	case define.QueryKindOfTerm:
		docQuery = f.createTermQuery(opt)
	case define.QueryKindOfPrefix:
		docQuery = f.createPrefixQuery(opt)
	case define.QueryKindOfMatchQuery:
		docQuery = f.createMatchQuery(opt)
	case define.QueryKindOfPhrase:
		docQuery = f.createPhraseQuery(opt)
	case define.QueryKindOfMatchAll:
		docQuery = bleve.NewMatchAllQuery()
	default:
		if opt.Key != "" {
			docQuery = f.createMatchQuery(opt)
		}else{
			docQuery = bleve.NewMatchAllQuery()
		}
	}

	//set filter fields
	//create bool query
	boolQuery := f.createFilterQuery(opt)
	if boolQuery != nil {
		if docQuery != nil {
			//add must doc query
			boolQuery.AddMust(docQuery)
		}
		//init multi condition search request
		searchRequest = bleve.NewSearchRequest(boolQuery)
	}else{
		//general search request
		searchRequest = bleve.NewSearchRequest(docQuery)
	}
	return searchRequest
}

////////////////////////////
//create filter bool query
////////////////////////////

func (f *Query) createFilterQuery(opt *json.QueryOptJson) *query.BooleanQuery {
	var (
		tempStr string
	)

	//check
	if opt.Filters == nil || len(opt.Filters) <= 0 {
		return nil
	}

	//init bool query
	boolQuery := bleve.NewBooleanQuery()

	//add filter field and value
	for _, filter := range opt.Filters {
		//do sub query by kind
		switch filter.Kind {
		case define.FilterKindBoolean:
			{
				//match by boolean
				boolVal, _ := filter.Val.(bool)
				pg := bleve.NewBoolFieldQuery(boolVal)
				pg.SetField(filter.Field)
				if filter.IsMust {
					boolQuery.AddMust(pg)
				}else{
					boolQuery.AddShould(pg)
				}
			}
		case define.FilterKindMatch:
			{
				//match by condition
				tempStr = fmt.Sprintf("%v", filter.Val)
				pg := bleve.NewTermQuery(tempStr)
				pg.SetField(filter.Field)
				if filter.IsMust {
					boolQuery.AddMust(pg)
				}else{
					boolQuery.AddShould(pg)
				}
			}
		case define.FilterKindMatchRange:
			{
				//match by range
				pg := bleve.NewTermRangeQuery(filter.MinVal, filter.MinVal)
				pg.SetField(filter.Field)
				if filter.IsMust {
					boolQuery.AddMust(pg)
				}else{
					boolQuery.AddShould(pg)
				}
			}
		case define.FilterKindPrefix:
			{
				pg := bleve.NewPrefixQuery(fmt.Sprintf("%v", filter.Val))
				pg.SetField(filter.Field)
				if filter.IsMust {
					boolQuery.AddMust(pg)
				}else{
					boolQuery.AddShould(pg)
				}
			}
		case define.FilterKindPhraseQuery, define.FilterKindExcludePhraseQuery:
			{
				//sub terms phrase query
				termSlice := make([]string, 0)
				switch filter.Val.(type) {
				case []string:
					termSlice = filter.Val.([]string)
				default:
					tmpStr := fmt.Sprintf("%v", filter.Val)
					termSlice = append(termSlice, tmpStr)
				}
				pq := bleve.NewPhraseQuery(termSlice, filter.Field)
				if filter.IsExclude {
					boolQuery.AddMustNot(pq)
				} else {
					if filter.IsMust {
						boolQuery.AddMust(pq)
					}else{
						boolQuery.AddShould(pq)
					}
				}
			}
		case define.FilterKindNumericRange:
			{
				//min <= val < max
				minFloatVal, _ := filter.MinFloatVal.Float64()
				maxFloatVal, _ := filter.MaxFloatVal.Float64()
				pg := bleve.NewNumericRangeQuery(&minFloatVal, &maxFloatVal)
				pg.SetField(filter.Field)
				if filter.IsMust {
					boolQuery.AddMust(pg)
				}else{
					boolQuery.AddShould(pg)
				}
			}
		case define.FilterKindDateRange:
			{
				pg := bleve.NewDateRangeQuery(filter.StartTime, filter.EndTime)
				pg.SetField(filter.Field)
				if filter.IsMust {
					boolQuery.AddMust(pg)
				}else{
					boolQuery.AddShould(pg)
				}
			}
		case define.FilterKindSubDocIds:
			{
				pg := bleve.NewDocIDQuery(filter.DocIds)
				if filter.IsMust {
					boolQuery.AddMust(pg)
				}else{
					boolQuery.AddShould(pg)
				}
			}
		case define.FilterKindTermsQuery:
			{
				for _, v := range filter.Terms {
					if v == "" {
						continue
					}
					pg := bleve.NewTermQuery(v)
					pg.SetField(filter.Field)
					if filter.IsMust {
						boolQuery.AddMust(pg)
					}else{
						boolQuery.AddShould(pg)
					}
				}
			}
		}
	}
	return boolQuery
}

//////////////////////
//create relate query
//////////////////////

func (f *Query) createPhraseQuery(opt *json.QueryOptJson) query.Query {
	subQuery := bleve.NewMatchPhraseQuery(opt.Key)
	if opt.Fields != nil {
		for _, field := range opt.Fields {
			//set query field
			subQuery.SetField(field)
		}
	}
	return subQuery
}

func (f *Query) createMatchQuery(opt *json.QueryOptJson) query.Query {
	subQuery := bleve.NewMatchQuery(opt.Key)
	if opt.Fields != nil {
		for _, field := range opt.Fields {
			//set query field
			subQuery.SetField(field)
		}
	}
	return subQuery
}

func (f *Query) createPrefixQuery(opt *json.QueryOptJson) query.Query {
	subQuery := bleve.NewPrefixQuery(opt.Key)
	return subQuery
}

func (f *Query) createTermQuery(opt *json.QueryOptJson) query.Query {
	subQuery := bleve.NewTermQuery(opt.TermPara.Val)
	subQuery.SetField(opt.TermPara.Field)
	return subQuery
}


///////////////
//private func
///////////////

//format result
func (f *Query) formatResult(
					index bleve.Index,
					hits *search.DocumentMatchCollection,
				) []*json.HitDocJson {
	//basic check
	if index == nil || hits == nil {
		return nil
	}

	//format result
	result := make([]*json.HitDocJson, 0)

	//format records
	for _, hit := range *hits {
		//get original doc
		doc, err := index.Document(hit.ID)
		if err != nil {
			continue
		}

		//analyze doc
		hitDocJson, err := f.AnalyzeDoc(doc, hit)
		if err != nil || hitDocJson == nil {
			continue
		}

		//add into slice
		result = append(result, hitDocJson)
	}
	return result
}
