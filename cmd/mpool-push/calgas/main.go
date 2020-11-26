package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"net/http"
)

func main() {
	var (
		addr      string = "http://172.30.9.77:1234/rpc/v0"
		authToken string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.PEhuw-RV-8xFVlIT2uJAtB7GyqMtJUnqJNO-vsyoO8o"
	)
	node, closer, err := NewFullNodeAPI(addr, authToken)
	if err != nil {
		panic(err)
	}
	defer closer()
	from, err := address.NewFromString("f1tyxygxzi6o6b3cf3wzte7rorrk4yy3ut4x2jxty")
	if err != nil {
		return
	}
	msg, err := node.GasEstimateMessageGas(context.Background(), &types.Message{
		From: from,
		To:   from,
	}, nil, types.EmptyTSK)
	if err != nil {
		panic(err)
	}
	act, err := node.StateGetActor(context.Background(), from, types.EmptyTSK)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		msg.Nonce = act.Nonce
	}

	fmt.Println("GasLimit", msg.GasLimit)
	fmt.Println("GasFeeCap", msg.GasFeeCap)
	fmt.Println("GasPremium", msg.GasPremium)
	fmt.Println("Nonce", msg.Nonce)
	fmt.Println(fmt.Sprintf("--from %v --gas-premium %v --gas-feecap %v --gas-limit %v --nonce %v", msg.From, msg.GasPremium, msg.GasFeeCap, msg.GasLimit, msg.Nonce))
}

func NewFullNodeAPI(addr, authToken string) (api.FullNode, jsonrpc.ClientCloser, error) {
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
	var fullNode apistruct.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), addr, "Filecoin", []interface{}{&fullNode.Internal, &fullNode.CommonStruct.Internal}, headers)
	return &fullNode, closer, err
}
