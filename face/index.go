package face

import (
	"errors"
	"fmt"
	"github.com/andyzhou/tinysearch/define"
	_ "github.com/andyzhou/tinysearch/jiebago/tokenizers" //for init tokenizers
	"github.com/blevesearch/bleve/v2"
	_ "github.com/blevesearch/bleve/v2/analysis/analyzer/custom" //for init 'custom'
	"github.com/blevesearch/bleve/v2/mapping"
	"os"
	"sync"
)

/*
 * face for index
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 * - chinese token base on 'github.com/wangbin/jiebago'
 */

//face info
type Index struct {
	indexDir string
	dictFile string
	tag      string
	indexer  bleve.Index
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
func (f *Index) RemoveIndex() error {
	//basic check
	if f.tag == "" {
		return errors.New("invalid tag")
	}
	err := os.RemoveAll(f.indexDir)
	return err
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
	index, subErr := bleve.New(subDir, indexMapping)
	if subErr != nil {
		//index had exists, open it.
		if subErr == bleve.ErrorIndexPathExists {
			index, subErr = bleve.Open(subDir)
		}
		if subErr != nil {
			return subErr
		}
	}

	//sync indexer
	f.Lock()
	defer f.Unlock()
	f.indexer = index
	return nil
}

//create index mapping
func (f *Index) CreateIndexMap(dictFile ...string) *mapping.IndexMappingImpl {
	if dictFile != nil {
		indexMapping, err := f.CreateChineseMap(dictFile[0])
		if err != nil {
			return nil
		}
		return indexMapping
	}
	indexMapping := mapping.NewIndexMapping()
	return indexMapping
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
		define.CustomTokenizerOfJieBa,
		map[string]interface{}{
			"file": dictPath,
			"type": define.CustomTokenizerOfJieBa,
		})
	if err != nil {
		return nil, err
	}

	// create a custom analyzer
	err = indexMapping.AddCustomAnalyzer(
		define.CustomTokenizerOfJieBa,
		map[string]interface{}{
			"type":      "custom",
			"tokenizer": define.CustomTokenizerOfJieBa,
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
	indexMapping.DefaultAnalyzer = define.CustomTokenizerOfJieBa
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