package api

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
)

type LotusGatewayAPI interface {
	// StateMinerInfo returns info about the indicated miner
	StateMinerInfo(context.Context, address.Address, types.TipSetKey) (miner.MinerInfo, error)
	// StateGetActor returns the indicated actor's nonce and balance.
	StateGetActor(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error)
	// WalletBalance returns the balance of the given address at the current head of the chain.
	WalletBalance(context.Context, address.Address) (types.BigInt, error)
}
