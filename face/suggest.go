package face

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/andyzhou/tinySearch/define"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	"github.com/blevesearch/bleve/v2"
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

//inter macro define
const (
	interSuggestIndexTag = "__suggester__"
	interSuggestChanSize = 1024
)

//face info
type Suggest struct {
	dataPath string
	index iface.IIndex
	syncReqChan chan json.SuggestJson
	closeChan chan bool
	Base
}

//construct
func NewSuggest(dataPath string) *Suggest {
	//self init
	this := &Suggest{
		dataPath:dataPath,
		syncReqChan:make(chan json.SuggestJson, interSuggestChanSize),
		closeChan:make(chan bool, 1),
	}
	this.interInit()
	go this.runMainProcess()
	return this
}

//quit
func (f *Suggest) Quit() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Suggest:Quit panic, err:", err)
		}
	}()
	f.closeChan <- true
}

//get suggest
func (f *Suggest) GetSuggest(
					opt *json.SuggestOptJson,
				) *json.SuggestsJson {
	//basic check
	if opt == nil {
		return nil
	}

	//get index
	indexer := f.index.GetIndex()
	if indexer == nil {
		return nil
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
	searchRequest.Size = define.RecPerPage
	searchRequest.Explain = true

	//begin search
	searchResult, err := indexer.Search(searchRequest)
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
	result.Total = int64(searchResult.Total)

	//format records
	for _, hit := range searchResult.Hits {
		//get original doc by id
		doc, err := indexer.Document(hit.ID)
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

	return result
}


//add new suggest
func (f *Suggest) AddSuggest(
					doc *json.SuggestJson,
				) (bRet bool) {
	//basic check
	if doc == nil {
		bRet = false
		return
	}

	defer func() {
		if err := recover(); err != nil {
			bRet = false
			log.Println("Suggest:AddSuggest panic, err:", err)
		}
	}()

	//send to chan
	f.syncReqChan <- *doc
	bRet = true
	return
}

//////////////
//private func
//////////////

//main process
func (f *Suggest) runMainProcess() {
	var (
		req json.SuggestJson
		isOk bool
	)

	defer func() {
		if err := recover(); err != nil {
			log.Println("Suggest:runMainProcess panic, err:", err)
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
					doc *json.SuggestJson,
				) bool {
	//basic check
	if doc == nil {
		return false
	}

	//get index
	indexer := f.index.GetIndex()
	if indexer == nil {
		return false
	}

	//add or update doc
	keyMd5 := f.genMd5(doc.Key)
	oldRec, err := indexer.Document(keyMd5)
	if err != nil {
		return false
	}
	if oldRec != nil {
		//analyze doc
		genMap := f.FormatDoc(oldRec)
		if genMap != nil {
			oldDocJson := json.NewSuggestJson()
			oldDocByte := oldDocJson.EncodeSimple(genMap)
			bRet := oldDocJson.Decode(oldDocByte)
			if bRet {
				//check doc count
				if oldDocJson.Count >= doc.Count {
					//same data, not need sync
					return false
				}
			}
		}
	}

	//sync into index
	err = indexer.Index(keyMd5, doc)
	if err != nil {
		log.Println("Suggest::AddSuggest failed, err:", err.Error())
		return false
	}
	return true
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

//inter init
func (f *Suggest) interInit() {
	//init inter suggest index
	f.index = NewIndex(f.dataPath, interSuggestIndexTag)

	//create index
	f.index.CreateIndex()
}