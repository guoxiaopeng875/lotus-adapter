package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/messagepool"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/config"
	"github.com/ipfs/go-cid"
	"golang.org/x/xerrors"
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
	if err := replace(node); err != nil {
		panic(err)
	}
}

// GasLimit 1563247970
//GasFeeCap 3008800
//GasPremium 2510326
func replace(node api.FullNode) error {
	ctx := context.Background()
	var from address.Address
	mcid, err := cid.Decode("bafy2bzaced44r5zrp3vqo7t5phl2r5q7aqgadsvyjwrb6hshqwlwbtkkdyfje")
	if err != nil {
		return err
	}

	msg, err := node.ChainGetMessage(ctx, mcid)
	if err != nil {
		return fmt.Errorf("could not find referenced message: %w", err)
	}

	from = msg.From

	ts, err := node.ChainHead(ctx)
	if err != nil {
		return xerrors.Errorf("getting chain head: %w", err)
	}

	pending, err := node.MpoolPending(ctx, ts.Key())
	if err != nil {
		return err
	}

	var found *types.SignedMessage
	for _, p := range pending {
		if p.Message.From == from {
			found = p
			break
		}
	}

	if found == nil {
		return fmt.Errorf("no pending message found from %s", from)
	}

	mmsg := found.Message

	minRBF := messagepool.ComputeMinRBF(msg.GasPremium)

	var mss *api.MessageSendSpec

	// msg.GasLimit = 0 // TODO: need to fix the way we estimate gas limits to account for the messages already being in the mempool
	msg.GasFeeCap = abi.NewTokenAmount(0)
	msg.GasPremium = abi.NewTokenAmount(0)
	retm, err := node.GasEstimateMessageGas(ctx, &mmsg, mss, types.EmptyTSK)
	if err != nil {
		return fmt.Errorf("failed to estimate gas values: %w", err)
	}

	msg.GasPremium = big.Max(retm.GasPremium, minRBF)
	msg.GasFeeCap = big.Max(retm.GasFeeCap, msg.GasPremium)

	mff := func() (abi.TokenAmount, error) {
		return abi.TokenAmount(config.DefaultDefaultMaxFee), nil
	}

	messagepool.CapGasFee(mff, &mmsg, mss.Get().MaxFee)
	fmt.Println("GasLimit", msg.GasLimit)
	fmt.Println("GasFeeCap", msg.GasFeeCap)
	fmt.Println("GasPremium", msg.GasPremium)
	return nil
}

func NewFullNodeAPI(addr, authToken string) (api.FullNode, jsonrpc.ClientCloser, error) {
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
	var fullNode apistruct.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), addr, "Filecoin", []interface{}{&fullNode.Internal, &fullNode.CommonStruct.Internal}, headers)
	return &fullNode, closer, err
}
