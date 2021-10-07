package face

import (
	genJson "encoding/json"
	"errors"
	"fmt"
	"github.com/andyzhou/tinySearch/define"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	"github.com/blevesearch/bleve/v2"
	"log"
)

/*
 * face for agg
 */

//face info
type Agg struct {
	query iface.IQuery //reference
}

//construct
func NewAgg(query iface.IQuery) *Agg {
	//self init
	this := &Agg{
		query: query,
	}
	return this
}

//get agg list
func (f *Agg) GetAggList(
				index iface.IIndex,
				opt *json.QueryOptJson,
			) (*json.AggregatesJson, error) {
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

	//default check
	if opt.AggSize <= 0 {
		opt.AggSize = define.RecPerPage
	}

	//build search request
	searchRequest := f.query.BuildSearchReq(opt)

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
	searchResult, err := indexer.Search(searchRequest)
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
			aggJson.Min = genJson.Number(fmt.Sprintf("%v", *v.Min))
			aggJson.Max = genJson.Number(fmt.Sprintf("%v", *v.Max))
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
