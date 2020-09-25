package iface

import (
	"context"
	search "github.com/andyzhou/tinySearch/pb"
)

/*
 * interface for rpc call back
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

type IRpcCB interface {
	DocRemove(
		ctx context.Context,
		in *search.DocRemoveReq,
	) (*search.DocSyncResp, error)

	DocSync(
		ctx context.Context,
		in *search.DocSyncReq,
	) (*search.DocSyncResp, error)
}

