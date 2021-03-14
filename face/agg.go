package face

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinySearch/define"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"log"
)

/*
 * face for agg
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Agg struct {
}

//construct
func NewAgg() *Agg {
	//self init
	this := &Agg{
	}
	return this
}

//get agg list
func (f *Agg) GetAggList(
				index iface.IIndex,
				opt *json.QueryOptJson,
			) (*json.AggregatesJson, error) {
	var (
		searchRequest *bleve.SearchRequest
		docQuery *query.MatchQuery
	)

	//basic check
	if index == nil || opt == nil {
		return nil, errors.New("invalid parameter")
	}

	if opt.Key == "" || opt.AggField == nil {
		return nil, errors.New("invalid parameter")
	}

	//get index
	indexer := index.GetIndex()
	if indexer == nil {
		return nil, errors.New("can't get indexer")
	}

	//init query
	docQuery = bleve.NewMatchQuery(opt.Key)

	if opt.AggSize <= 0 {
		opt.AggSize = define.RecPerPage
	}

	//general search request
	searchRequest = bleve.NewSearchRequest(docQuery)

	//set aggregating facet
	aggField := opt.AggField
	facetName := aggField.Field
	facet := bleve.NewFacetRequest(aggField.Field, aggField.Size)
	if aggField.IsNumeric {
		//numeric
		for _, numeric := range aggField.NumericRanges {
			name := fmt.Sprintf("%s-%d", facetName, int(numeric.From))
			facet.AddNumericRange(name, &numeric.From, &numeric.To)
		}
	}

	//add sub facet into search request
	searchRequest.AddFacet(facetName, facet)

	//begin search
	searchResult, err := (*indexer).Search(searchRequest)
	if err != nil {
		log.Println("Agg::GetAggList failed, err:", err.Error())
		return nil, err
	}

	//format facet result
	facetResult, ok := searchResult.Facets[facetName]
	if !ok || facetResult == nil {
		return nil, nil
	}

	//init final result
	result := json.NewAggregatesJson()

	if aggField.IsNumeric {
		//numeric
		for _, v := range facetResult.NumericRanges {
			//format final query for agg
			aggJson := json.NewAggregateJson()
			aggJson.Name = v.Name
			aggJson.Min = *v.Min
			aggJson.Max = *v.Max
			aggJson.Count = v.Count

			//add into slice
			result.AddObj(aggJson)
		}
	}else{
		//term
		for _, v := range facetResult.Terms {
			//format final query for agg
			aggJson := json.NewAggregateJson()
			aggJson.Name = v.Term
			aggJson.Count = v.Count

			//add into slice
			result.AddObj(aggJson)
		}
	}

	return result, nil
}
