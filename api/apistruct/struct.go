package apistruct

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	api2 "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
	"github.com/guoxiaopeng875/lotus-adapter/api"
	"github.com/guoxiaopeng875/lotus-adapter/api/apitypes"
)

type LotusGatewayStruct struct {
	Internal struct {
		StateMinerInfo func(ctx context.Context, address address.Address, key types.TipSetKey) (miner.MinerInfo, error)
		StateGetActor  func(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error)
		WalletBalance  func(ctx context.Context, address address.Address) (types.BigInt, error)
		MinerAssetInfo func(ctx context.Context, miner address.Address) (*apitypes.ClusterAssetInfo, error)
		WorkerJobs     func(ctx context.Context) (map[uint64][]storiface.WorkerJob, error)
		SectorsList    func(ctx context.Context) ([]abi.SectorNumber, error)
		WorkerStats    func(ctx context.Context) (map[uint64]storiface.WorkerStats, error)
		SectorsStatus  func(ctx context.Context, sid abi.SectorNumber, showOnChainInfo bool) (api2.SectorInfo, error)
	}
}

func (l *LotusGatewayStruct) SectorsStatus(ctx context.Context, sid abi.SectorNumber, showOnChainInfo bool) (api2.SectorInfo, error) {
	return l.Internal.SectorsStatus(ctx, sid, showOnChainInfo)
}

func (l *LotusGatewayStruct) WorkerStats(ctx context.Context) (map[uint64]storiface.WorkerStats, error) {
	return l.Internal.WorkerStats(ctx)
}

func (l *LotusGatewayStruct) SectorsList(ctx context.Context) ([]abi.SectorNumber, error) {
	return l.Internal.SectorsList(ctx)
}

func (l *LotusGatewayStruct) WorkerJobs(ctx context.Context) (map[uint64][]storiface.WorkerJob, error) {
	return l.Internal.WorkerJobs(ctx)
}

func (l *LotusGatewayStruct) StateMinerInfo(ctx context.Context, a address.Address, key types.TipSetKey) (miner.MinerInfo, error) {
	return l.Internal.StateMinerInfo(ctx, a, key)
}

func (l *LotusGatewayStruct) StateGetActor(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error) {
	return l.Internal.StateGetActor(ctx, actor, tsk)
}

func (l *LotusGatewayStruct) WalletBalance(ctx context.Context, a address.Address) (types.BigInt, error) {
	return l.Internal.WalletBalance(ctx, a)
}

func (l *LotusGatewayStruct) MinerAssetInfo(ctx context.Context, miner address.Address) (*apitypes.ClusterAssetInfo, error) {
	return l.Internal.MinerAssetInfo(ctx, miner)
}

var _ api.LotusGatewayAPI = &LotusGatewayStruct{}
