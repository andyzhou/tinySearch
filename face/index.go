package face

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"log"
	"os"
	"sync"
)

/*
 * face for index
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type Index struct {
	indexDir string
	tag string
	indexer *bleve.Index
	sync.RWMutex
}

//construct
func NewIndex(indexDir, tag string) *Index {
	//self init
	this := &Index{
		indexDir:indexDir,
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
func (f *Index) GetIndex() *bleve.Index {
	//basic check
	if f.tag == "" || f.indexer == nil {
		return nil
	}
	return f.indexer
}

//create index
func (f *Index) CreateIndex() error {
	//basic check
	if f.tag == "" || f.indexer != nil {
		return errors.New("invalid parameter")
	}

	//init index mapping
	indexMapping := mapping.NewIndexMapping()

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
	f.indexer = &index

	return nil
}
