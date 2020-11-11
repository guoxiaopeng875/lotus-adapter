package api

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
	"github.com/guoxiaopeng875/lotus-adapter/api/apitypes"
)

type LotusGatewayAPI interface {
	// StateMinerInfo returns info about the indicated miner
	StateMinerInfo(context.Context, address.Address, types.TipSetKey) (miner.MinerInfo, error)
	// StateGetActor returns the indicated actor's nonce and balance.
	StateGetActor(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error)
	// WalletBalance returns the balance of the given address at the current head of the chain.
	WalletBalance(context.Context, address.Address) (types.BigInt, error)
	// MinerAssetInfo
	MinerAssetInfo(ctx context.Context, miner address.Address) (*apitypes.ClusterAssetInfo, error)
	// -----------minerAPI-------------
	// WorkerJobs
	WorkerJobs(context.Context) (map[uint64][]storiface.WorkerJob, error)
	// SectorsList
	SectorsList(ctx context.Context) ([]abi.SectorNumber, error)
	// WorkerStats
	WorkerStats(context.Context) (map[uint64]storiface.WorkerStats, error)
	// SectorsStatus
	SectorsStatus(ctx context.Context, sid abi.SectorNumber, showOnChainInfo bool) (api.SectorInfo, error)
}
