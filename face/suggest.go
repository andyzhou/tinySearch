package face

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/json"
	"github.com/andyzhou/tinysearch/lib"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"log"
)

/*
 * face for suggest
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//suggest record field
const (
	SuggestFieldKey = "key"
	SuggestFieldCount = "count"
)

//inter data
type (
	suggestDocSync struct {
		indexTag string
		doc json.SuggestJson
	}
)

//face info
type Suggest struct {
	manager iface.IManager //parent reference
	queue *lib.Queue
	Base
}

//construct
func NewSuggest(manager iface.IManager) *Suggest {
	//self init
	this := &Suggest{
		manager: manager,
		queue: lib.NewQueue(),
	}
	this.interInit()
	return this
}

//quit
func (f *Suggest) Quit() {
	if f.queue != nil {
		f.queue.Quit()
	}
}

//get suggest, sort by count desc
//include get top hot keys, set `key` field as empty
func (f *Suggest) GetSuggest(
		opt *json.SuggestOptJson,
	) (*json.SuggestsJson, error) {
	var (
		docQuery query.Query
	)
	//basic check
	if opt == nil || opt.IndexTag == "" {
		return nil, errors.New("invalid parameter")
	}

	//get index
	indexer := f.getIndex(opt.IndexTag)
	if indexer == nil {
		return nil, errors.New("invalid index tag")
	}
	if opt.Page <= 0 {
		opt.Page = 1
	}
	if opt.PageSize <= 0 {
		opt.PageSize = define.RecPerPage
	}

	//init query by kind
	switch opt.QueryKind {
	case define.QueryKindOfPhrase:
		docQuery = bleve.NewMatchPhraseQuery(opt.Key)
	case define.QueryKindOfPrefix:
		docQuery = bleve.NewPrefixQuery(opt.Key)
	case define.QueryKindOfMatchQuery:
		docQuery = bleve.NewMatchQuery(opt.Key)
	default:
		docQuery = bleve.NewMatchAllQuery()
	}

	//set query field
	//docQuery.SetField("key")

	//set filter field
	//init bool query
	boolQuery := bleve.NewBooleanQuery()

	//add must query
	boolQuery.AddMust(docQuery)

	//init multi condition search request
	searchRequest := bleve.NewSearchRequest(boolQuery)

	//check and set query field
	if opt.Key != "" {
		searchRequest.Fields = []string{SuggestFieldKey}
	}

	//set sort by count desc
	customSort := make([]search.SearchSort, 0)
	cs := search.SortField{
		Field: SuggestFieldCount,
		Desc: true,
	}
	customSort = append(customSort, &cs)
	searchRequest.SortByCustom(customSort)

	//set others
	searchRequest.From = (opt.Page - 1) * opt.PageSize
	searchRequest.Size = opt.PageSize
	searchRequest.Explain = true

	//begin search
	searchResult, err := indexer.GetIndex().Search(searchRequest)
	if err != nil {
		return nil, err
	}

	//check hits
	if searchResult.Hits == nil ||
		searchResult.Hits.Len() <= 0 {
		return nil, nil
	}

	//init result
	result := json.NewSuggestsJson()
	result.Total = int64(searchResult.Total)

	//format records
	for _, hit := range searchResult.Hits {
		//get original doc by id
		doc, err := indexer.GetIndex().Document(hit.ID)
		if err != nil {
			continue
		}

		//init doc json
		suggestJson := json.NewSuggestJson()

		//format fields
		genMap := f.FormatDoc(doc)
		for k, v := range genMap {
			switch k {
			case SuggestFieldKey:
				{
					v1, ok := v.(string)
					if ok {
						suggestJson.Key = v1
					}
				}
			case SuggestFieldCount:
				{
					v1, ok := v.(float64)
					if ok {
						suggestJson.Count = int64(v1)
					}
				}
			}
		}
		//add into slice
		result.AddObj(suggestJson)
	}
	return result, nil
}

//add new suggest
func (f *Suggest) AddSuggest(
		indexTag string,
		doc *json.SuggestJson,
	) error {
	var (
		m any = nil
	)
	//basic check
	if indexTag == "" || doc == nil {
		return errors.New("invalid parameter")
	}

	//check index tag is register or not
	if f.getIndex(indexTag) == nil {
		return errors.New("can't get indexer by tag")
	}

	defer func() {
		if err := recover(); err != m {
			log.Printf("tinysearch.Suggest:AddSuggest panic, err:%v", err)
		}
	}()

	//init sync doc
	syncDoc := suggestDocSync{
		indexTag: indexTag,
		doc: *doc,
	}

	//send to queue
	_, err := f.queue.SendData(syncDoc)
	return err
}

//register suggest index
func (f *Suggest) RegisterSuggest(
	tags ...string) error {
	var (
		indexName string
		err error
	)
	//check
	if tags == nil || len(tags) <= 0 {
		return errors.New("invalid parameter")
	}
	//add suggest index names
	for _, tag := range tags {
		indexName = f.getIndexName(tag)
		err = f.manager.AddIndex(indexName)
	}
	return err
}

//////////////
//private func
//////////////

//cb for queue opt
func (f *Suggest) cbForQueueOpt(
	input interface{}) (interface{}, error) {
	//check
	if input == nil {
		return nil, errors.New("invalid parameter")
	}
	req, ok := input.(suggestDocSync)
	if !ok || &req == nil {
		return nil, errors.New("invalid request data format")
	}

	//process add suggest
	err := f.addSuggestProcess(&req)
	return nil, err
}

//inter init
func (f *Suggest) interInit() {
	f.queue.SetCallback(f.cbForQueueOpt)
}

//process add suggest request
func (f *Suggest) addSuggestProcess(
		req *suggestDocSync,
	) error {
	//basic check
	if req == nil {
		return errors.New("invalid parameter")
	}

	//get index
	indexer := f.getIndex(req.indexTag)
	if indexer == nil {
		return errors.New("can't get indexer by tag")
	}

	//add or update doc
	keyMd5 := f.genMd5(req.doc.Key)
	oldRec, err := indexer.GetIndex().Document(keyMd5)
	if err != nil {
		return err
	}
	if oldRec != nil {
		//analyze doc
		genMap := f.FormatDoc(oldRec)
		if genMap != nil {
			oldDocJson := json.NewSuggestJson()
			oldDocByte, err := oldDocJson.EncodeSimple(genMap)
			if err != nil {
				return err
			}
			err = oldDocJson.Decode(oldDocByte)
			if err == nil {
				//check doc count
				if oldDocJson.Count >= req.doc.Count {
					//same data, not need sync
					return errors.New("same data, not need sync")
				}
			}
		}
	}

	//sync into index
	err = indexer.GetIndex().Index(keyMd5, req.doc)
	return err
}

//get index by tag
func (f *Suggest) getIndex(tag string) iface.IIndex {
	//add suggest index name
	indexName := f.getIndexName(tag)
	index := f.manager.GetIndex(indexName)
	return index
}

//generate md5 value
func (f *Suggest) genMd5(orgString string) string {
	if len(orgString) <= 0 {
		return ""
	}
	m := md5.New()
	m.Write([]byte(orgString))
	return hex.EncodeToString(m.Sum(nil))
}

//get suggest index name
func (f *Suggest) getIndexName(tag string) string {
	return fmt.Sprintf(define.InterSuggestIndexPara, tag)
}