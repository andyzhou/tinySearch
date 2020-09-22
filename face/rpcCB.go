package face

import (
	"context"
	"errors"
	"fmt"
	"github.com/andyzhou/tinySearch/iface"
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

	//add into local index
	indexer := index.GetIndex()
	err := (*indexer).Index(in.DocId, in.Json)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	//format result
	result := &search.DocSyncResp{
		Success:true,
	}
	return result, nil
}
