package face

import (
	genJson "encoding/json"
	"errors"
	"fmt"
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/json"
	"github.com/blevesearch/bleve/v2"
)

/*
 * face for agg
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
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

	if opt.Key == "" || opt.AggFields == nil {
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

	//add batch aggregating facet
	tempAggFieldMap := map[string]*json.AggField{}
	for _, aggField := range opt.AggFields {
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
		tempAggFieldMap[facetName] = aggField
	}

	//begin search
	searchResult, err := indexer.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if searchResult.Facets == nil || len(searchResult.Facets) <= 0 {
		return nil, nil
	}

	//init final result
	result := json.NewAggregatesJson()

	//format facet result
	for facetName, facetResult := range searchResult.Facets {
		//check
		if facetName == "" || facetResult == nil {
			continue
		}
		aggField, ok := tempAggFieldMap[facetName]
		if !ok || aggField == nil {
			continue
		}

		//analyze one agg facet result
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
				result.AddObj(facetName, aggJson)
			}
		}else{
			//term
			if facetResult.Terms != nil {
				for _, v := range facetResult.Terms.Terms() {
					//format final query for agg
					aggJson := json.NewAggregateJson()
					aggJson.Name = v.Term
					aggJson.Count = v.Count
					//add into slice
					result.AddObj(facetName, aggJson)
				}
			}
		}
	}
	return result, nil
}
