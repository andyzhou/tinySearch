package face

import (
	"context"
	"errors"
	"fmt"
	"github.com/andyzhou/tinySearch/iface"
	"github.com/andyzhou/tinySearch/json"
	search "github.com/andyzhou/tinySearch/pb"
)

/*
 * face for rpc call back
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//face info
type IRpcCB struct {
	manager iface.IManager //manager reference
	json.BaseJson
}

//construct
func NewIRpcCB(
			manager iface.IManager,
		) *IRpcCB {
	//self init
	this := &IRpcCB{
		manager:manager,
	}
	return this
}

//////////////////////
//call backs for rpc
/////////////////////

//doc remove
func (f *IRpcCB) DocRemove(
					ctx context.Context,
					in *search.DocRemoveReq,
				) (*search.DocSyncResp, error) {
	var (
		tip string
	)

	//check input value
	if in == nil {
		return nil, errors.New("invalid parameter")
	}

	//get index
	index := f.manager.GetIndex(in.Tag)
	if index == nil {
		tip = fmt.Sprintf("can't get index by tag of %s", in.Tag)
		return nil, errors.New(tip)
	}

	//remove from local index
	indexer := index.GetIndex()
	err := (*indexer).Delete(in.DocId)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	//format result
	result := &search.DocSyncResp{
		Success:true,
	}
	return result, nil
}

//doc sync
func (f *IRpcCB) DocSync(
					ctx context.Context,
					in *search.DocSyncReq,
				) (*search.DocSyncResp, error) {
	var (
		tip string
	)

	//check input value
	if in == nil {
		return nil, errors.New("invalid parameter")
	}

	//get index
	index := f.manager.GetIndex(in.Tag)
	if index == nil {
		tip = fmt.Sprintf("can't get index by tag of %s", in.Tag)
		return nil, errors.New(tip)
	}

	//decode json byte
	kvMap := make(map[string]interface{})
	bRet := f.BaseJson.DecodeSimple(in.Json, kvMap)
	if !bRet {
		return nil, errors.New("decode json byte failed")
	}

	//add into local index
	indexer := index.GetIndex()
	err := (*indexer).Index(in.DocId, kvMap)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	//format result
	result := &search.DocSyncResp{
		Success:true,
	}
	return result, nil
}
