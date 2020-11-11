package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apibstore"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
	"github.com/filecoin-project/lotus/lib/blockstore"
	"github.com/filecoin-project/lotus/lib/bufbstore"
	"github.com/filecoin-project/lotus/storage"
	"github.com/guoxiaopeng875/lotus-adapter/api/apitypes"
	cbor "github.com/ipfs/go-ipld-cbor"
	"github.com/patrickmn/go-cache"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
)

func NewCachedFullNode(under api.FullNode, cache *cache.Cache) *CachedFullNode {
	return &CachedFullNode{nodeApi: under, cache: cache}
}

type CachedFullNode struct {
	nodeApi  api.FullNode
	minerApi api.StorageMiner
	cache    *cache.Cache
}

func (c *CachedFullNode) WorkerStats(ctx context.Context) (map[uint64]storiface.WorkerStats, error) {
	k := fmt.Sprintf("WorkerStats")
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(map[uint64]storiface.WorkerStats), nil
	}
	info, err := c.minerApi.WorkerStats(ctx)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(k, info)
	return info, nil
}

func (c *CachedFullNode) SectorsList(ctx context.Context) ([]abi.SectorNumber, error) {
	k := fmt.Sprintf("SectorsList")
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.([]abi.SectorNumber), nil
	}
	info, err := c.minerApi.SectorsList(ctx)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(k, info)
	return info, nil
}

func (c *CachedFullNode) WorkerJobs(ctx context.Context) (map[uint64][]storiface.WorkerJob, error) {
	k := fmt.Sprintf("WorkerJobs")
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(map[uint64][]storiface.WorkerJob), nil
	}
	info, err := c.minerApi.WorkerJobs(ctx)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(k, info)
	return info, nil
}

func (c *CachedFullNode) MinerAssetInfo(ctx context.Context, mAddr address.Address) (*apitypes.ClusterAssetInfo, error) {
	k := fmt.Sprintf("MinerAssetInfo:%s", mAddr.String())
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(*apitypes.ClusterAssetInfo), nil
	}
	info, err := c.minerAssetInfo(ctx, mAddr)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(k, info)
	return info, nil
}

func (c *CachedFullNode) minerAssetInfo(ctx context.Context, mAddr address.Address) (*apitypes.ClusterAssetInfo, error) {
	mi, err := c.nodeApi.StateMinerInfo(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}
	mAct, err := c.nodeApi.StateGetActor(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}
	tbs := bufbstore.NewTieredBstore(apibstore.NewAPIBlockstore(c.nodeApi), blockstore.NewTemporary())
	mas, err := miner.Load(adt.WrapStore(ctx, cbor.NewCborStore(tbs)), mAct)
	if err != nil {
		return nil, err
	}
	lockedFunds, err := mas.LockedFunds()
	if err != nil {
		return nil, err
	}
	availBalance, err := mas.AvailableBalance(mAct.Balance)
	if err != nil {
		return nil, err
	}
	postAddr, err := storage.AddressFor(ctx, c.nodeApi, mi, storage.PoStAddr, types.FromFil(1))
	if err != nil {
		return nil, xerrors.Errorf("getting address for post: %w", err)
	}
	postBls, err := c.nodeApi.WalletBalance(ctx, postAddr)
	if err != nil {
		return nil, err
	}
	return &apitypes.ClusterAssetInfo{
		MinerID:                  mAddr.String(),
		MinerBalance:             mAct.Balance,
		VestingFunds:             lockedFunds.VestingFunds,
		InitialPledgeRequirement: lockedFunds.InitialPledgeRequirement,
		PreCommitDeposits:        lockedFunds.PreCommitDeposits,
		AvailableBalance:         availBalance,
		PostBalance:              postBls,
	}, nil
}

func (c *CachedFullNode) WalletBalance(ctx context.Context, address address.Address) (types.BigInt, error) {
	k := fmt.Sprintf("WalletBalance:%s", address.String())
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(types.BigInt), nil
	}
	wb, err := c.nodeApi.WalletBalance(ctx, address)
	if err != nil {
		return types.EmptyInt, err
	}
	c.cache.SetDefault(k, wb)
	return wb, nil
}

func (c *CachedFullNode) StateGetActor(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error) {
	k := fmt.Sprintf("StateGetActor:%s", actor.String())
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(*types.Actor), nil
	}
	act, err := c.nodeApi.StateGetActor(ctx, actor, tsk)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(k, act)
	return act, nil
}

func (c *CachedFullNode) StateMinerInfo(ctx context.Context, address address.Address, key types.TipSetKey) (miner.MinerInfo, error) {
	k := fmt.Sprintf("StateMinerInfo:%s", address.String())
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(miner.MinerInfo), nil
	}
	mi, err := c.nodeApi.StateMinerInfo(ctx, address, key)
	if err != nil {
		return miner.MinerInfo{}, err
	}
	c.cache.SetDefault(k, mi)
	return mi, nil
}
