package apistruct

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/guoxiaopeng875/lotus-adapter/api"
)

type LotusGatewayStruct struct {
	Internal struct {
		StateMinerInfo func(ctx context.Context, address address.Address, key types.TipSetKey) (miner.MinerInfo, error)
		StateGetActor  func(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error)
		WalletBalance  func(ctx context.Context, address address.Address) (types.BigInt, error)
	}
}

func (l *LotusGatewayStruct) StateMinerInfo(ctx context.Context, address address.Address, key types.TipSetKey) (miner.MinerInfo, error) {
	panic("implement me")
}

func (l *LotusGatewayStruct) StateGetActor(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error) {
	panic("implement me")
}

func (l *LotusGatewayStruct) WalletBalance(ctx context.Context, address address.Address) (types.BigInt, error) {
	panic("implement me")
}

var _ api.LotusGatewayAPI = &LotusGatewayStruct{}
