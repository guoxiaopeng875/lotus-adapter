package client

import (
	"context"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/guoxiaopeng875/lotus-adapter/api"
	"github.com/guoxiaopeng875/lotus-adapter/api/apistruct"
	"net/http"
)

// NewLotusGatewayRPC creates a new http jsonrpc client for lotus
func NewLotusGatewayRPC(ctx context.Context, addr string, requestHeader http.Header, opts ...jsonrpc.Option) (api.LotusGatewayAPI, jsonrpc.ClientCloser, error) {
	var res apistruct.LotusGatewayStruct
	closer, err := jsonrpc.NewMergeClient(ctx, addr, "Filecoin",
		[]interface{}{
			&res.Internal,
		},
		requestHeader,
		opts...,
	)

	return &res, closer, err
}
