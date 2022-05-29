package face

import (
	"errors"
	"fmt"
	_ "github.com/andyzhou/tinySearch/jiebago/tokenizers" //for init tokenizers
	"github.com/blevesearch/bleve/v2"
	_ "github.com/blevesearch/bleve/v2/analysis/analyzer/custom" //for init 'custom'
	"github.com/blevesearch/bleve/v2/mapping"
	"log"
	"os"
	"sync"
)

/*
 * face for index
 * - chinese token base on 'github.com/wangbin/jiebago'
 */

//inter macro define
const (
	CustomTokenizerOfJieBa = "jieba"
)

//face info
type Index struct {
	indexDir string
	dictFile string
	tag string
	indexer bleve.Index
	sync.RWMutex
}

//construct
func NewIndex(indexDir, tag string, dictFile ...string) *Index {
	var (
		dictFilePath string
	)
	if dictFile != nil && dictFile[0] != "" {
		dictFilePath = dictFile[0]
	}
	//self init
	this := &Index{
		indexDir:indexDir,
		dictFile: dictFilePath,
		tag:tag,
	}
	return this
}

//remove index
func (f *Index) RemoveIndex() bool {
	//basic check
	if f.tag == "" {
		return false
	}
	err := os.RemoveAll(f.indexDir)
	if err != nil {
		log.Println("Index::RemoveIndex failed, err:", err.Error())
		return false
	}
	return true
}

//get index
func (f *Index) GetIndex() bleve.Index {
	//basic check
	if f.tag == "" || f.indexer == nil {
		return nil
	}
	return f.indexer
}

//create index
func (f *Index) CreateIndex() error {
	var (
		indexMapping *mapping.IndexMappingImpl
		err error
	)

	//basic check
	if f.tag == "" || f.indexer != nil {
		return errors.New("invalid parameter")
	}

	if f.dictFile != "" {
		//create index with chinese tokenizer support
		indexMapping, err = f.CreateChineseMap(f.dictFile)
		if err != nil {
			return err
		}
	}else{
		//init default index mapping
		indexMapping = mapping.NewIndexMapping()
	}

	//format sub dir path
	subDir := fmt.Sprintf("%s/%s", f.indexDir, f.tag)

	//init search index
	index, err := bleve.New(subDir, indexMapping)
	if err != nil {
		//index had exists, open it.
		if err == bleve.ErrorIndexPathExists {
			index, err = bleve.Open(subDir)
		}
		if err != nil {
			log.Println("Index::CreateIndex failed, err:", err.Error())
			return err
		}
	}

	//sync indexer
	f.Lock()
	defer f.Unlock()
	f.indexer = index

	return nil
}

//create chinese index mapping
func (f *Index) CreateChineseMap(dictPath string) (*mapping.IndexMappingImpl, error) {
	if dictPath == "" {
		return nil, errors.New("invalid dict path for index")
	}

	// open a new index
	indexMapping := bleve.NewIndexMapping()

	//set tokenizer
	err := indexMapping.AddCustomTokenizer(
		CustomTokenizerOfJieBa,
		map[string]interface{}{
			"file": dictPath,
			"type": CustomTokenizerOfJieBa,
		})
	if err != nil {
		return nil, err
	}

	// create a custom analyzer
	err = indexMapping.AddCustomAnalyzer(
		CustomTokenizerOfJieBa,
		map[string]interface{}{
			"type":      "custom",
			"tokenizer": CustomTokenizerOfJieBa,
			"token_filters": []string{
				"possessive_en",
				"to_lower",
				"stop_en",
			},
		})

	if err != nil {
		return nil, err
	}

	//set default analyzer
	indexMapping.DefaultAnalyzer = CustomTokenizerOfJieBa
	return indexMapping, nil
}

//set tokenizer file
func (f *Index) SetDictPath(dict string) bool {
	if dict == "" {
		return false
	}
	f.dictFile = dict
	return true
}