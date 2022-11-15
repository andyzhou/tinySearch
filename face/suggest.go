package face

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/andyzhou/tinysearch/define"
	"github.com/andyzhou/tinysearch/iface"
	"github.com/andyzhou/tinysearch/json"
	"github.com/blevesearch/bleve/v2"
	"log"
)

/*
 * face for suggest
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
	syncReqChan chan suggestDocSync
	closeChan chan struct{}
	Base
}

//construct
func NewSuggest(manager iface.IManager) *Suggest {
	//self init
	this := &Suggest{
		manager: manager,
		syncReqChan:make(chan suggestDocSync, define.InterSuggestChanSize),
		closeChan:make(chan struct{}, 1),
	}
	go this.runMainProcess()
	return this
}

//quit
func (f *Suggest) Quit() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("tinysearch.Suggest:Quit panic, err:%v", err)
		}
	}()
	f.closeChan <- struct{}{}
}

//get suggest
func (f *Suggest) GetSuggest(
					opt *json.SuggestOptJson,
				) (*json.SuggestsJson, error) {
	//basic check
	if opt == nil || opt.IndexTag == "" || opt.Key == "" {
		return nil, errors.New("invalid parameter")
	}

	//get index
	indexer := f.getIndex(opt.IndexTag)
	if indexer == nil {
		return nil, errors.New("invalid index tag")
	}
	if opt.Size <= 0 {
		opt.Size = define.RecPerPage
	}

	//init query
	docQuery := bleve.NewPrefixQuery(opt.Key)

	//set query field
	docQuery.SetField("key")

	//set filter field
	//init bool query
	boolQuery := bleve.NewBooleanQuery()

	//add must query
	boolQuery.AddMust(docQuery)

	//init multi condition search request
	searchRequest := bleve.NewSearchRequest(boolQuery)

	//set others
	searchRequest.From = 0
	searchRequest.Size = opt.Size
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
	//basic check
	if indexTag == "" || doc == nil {
		return errors.New("invalid parameter")
	}

	//check index tag is register or not
	if f.getIndex(indexTag) == nil {
		return errors.New("can't get indexer by tag")
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("tinysearch.Suggest:AddSuggest panic, err:%v", err)
		}
	}()

	//init sync doc
	syncDoc := suggestDocSync{
		indexTag: indexTag,
		doc: *doc,
	}

	//send to chan
	select {
	case f.syncReqChan <- syncDoc:
	}
	return nil
}

//register suggest index
func (f *Suggest) RegisterSuggest(tag string) error {
	//check
	if tag == "" {
		return errors.New("invalid parameter")
	}
	//add suggest index name
	indexName := f.getIndexName(tag)
	err := f.manager.AddIndex(indexName)
	return err
}

//////////////
//private func
//////////////

//main process
func (f *Suggest) runMainProcess() {
	var (
		req suggestDocSync
		isOk bool
	)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("tinysearch.Suggest:runMainProcess panic, err:%v", err)
		}
		close(f.syncReqChan)
		close(f.closeChan)
	}()

	//loop
	for {
		select {
		case req, isOk = <- f.syncReqChan:
			if isOk {
				f.addSuggestProcess(&req)
			}
		case <- f.closeChan:
			return
		}
	}
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

//get suggest index name
func (f *Suggest) getIndexName(tag string) string {
	return fmt.Sprintf(define.InterSuggestIndexPara, tag)
}