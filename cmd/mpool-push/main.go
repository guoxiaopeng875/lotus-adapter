package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/builtin/multisig"
	"github.com/ipfs/go-cid"
	"net/http"
)

func main() {
	var (
		msgHex    string = "828a005502a1746faf1e35c9e5f6ee8ac3d9c82eb070fcc8a855019e2f835f28f3bc1d88bbb6664fc5d18ab98c6e9304401a02bffea844000f74cc44000f4ba003438200405842016ed7b74acc97f0c49af2e769fd1ea40cc2724d88cf4c8d835b32e385c899f4bc04db70b8904093f804a25c5bdb0a28f7561c6539445443cb8da983ae53cb487b00"
		addr      string = "http://172.30.9.77:1234/rpc/v0"
		authToken string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.PEhuw-RV-8xFVlIT2uJAtB7GyqMtJUnqJNO-vsyoO8o"
	)
	ctx := context.Background()
	api, closer, err := NewFullNodeAPI(addr, authToken)
	if err != nil {
		panic(err)
	}
	defer closer()
	msgBytes, err := hex.DecodeString(msgHex)
	if err != nil {
		panic(err)
	}
	sm, err := types.DecodeSignedMessage(msgBytes)
	if err != nil {
		panic(err)
	}
	cid, err := api.MpoolPush(ctx, sm)
	if err != nil {
		panic(err)
	}
	fmt.Println(cid.String())

	if err := checkMessage(api, cid); err != nil {
		panic(err)
	}
}

func checkMessage(api api.FullNode, mCid cid.Cid) error {
	ctx := context.Background()
	wait, err := api.StateWaitMsg(ctx, mCid, 0)
	if err != nil {
		return err
	}

	if wait.Receipt.ExitCode != 0 {
		return fmt.Errorf("proposal returned exit %d", wait.Receipt.ExitCode)
	}

	var retval multisig.ProposeReturn
	if err := retval.UnmarshalCBOR(bytes.NewReader(wait.Receipt.Return)); err != nil {
		return fmt.Errorf("failed to unmarshal propose return value: %w", err)
	}

	fmt.Printf("Transaction ID: %d\n", retval.TxnID)
	if retval.Applied {
		fmt.Printf("Transaction was executed during propose\n")
		fmt.Printf("Exit Code: %d\n", retval.Code)
		fmt.Printf("Return Value: %x\n", retval.Ret)
	}

	return nil
}

func NewFullNodeAPI(addr, authToken string) (api.FullNode, jsonrpc.ClientCloser, error) {
	headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
	var fullNode apistruct.FullNodeStruct
	closer, err := jsonrpc.NewMergeClient(context.Background(), addr, "Filecoin", []interface{}{&fullNode.Internal, &fullNode.CommonStruct.Internal}, headers)
	return &fullNode, closer, err
}
