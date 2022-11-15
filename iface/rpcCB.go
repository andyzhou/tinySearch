package iface

import (
	"context"
	search "github.com/andyzhou/tinysearch/pb"
)

/*
 * interface for rpc service call back
 */

type IRpcCB interface {
	DocQuery(
		ctx context.Context,
		in *search.DocQueryReq,
	) (*search.DocQueryResp, error)

	DocRemove(
		ctx context.Context,
		in *search.DocRemoveReq,
	) (*search.DocSyncResp, error)

	DocGet(
		ctx context.Context,
		in *search.DocGetReq,
	) (*search.DocGetResp, error)

	DocSync(
		ctx context.Context,
		in *search.DocSyncReq,
	) (*search.DocSyncResp, error)
}

