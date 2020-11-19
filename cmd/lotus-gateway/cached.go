package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/google/uuid"
	"github.com/guoxiaopeng875/lotus-adapter/api/apitypes"
	"github.com/guoxiaopeng875/lotus-adapter/apiwrapper"
	"github.com/patrickmn/go-cache"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
)

func NewCachedFullNode(nodeApi api.FullNode, minerApi api.StorageMiner, cache *cache.Cache, secret *dtypes.APIAlg) *CachedFullNode {
	return &CachedFullNode{nodeApi: nodeApi, minerApi: minerApi, cache: cache, APISecret: secret,
		wrapper: apiwrapper.NewLotusAPIWrapper(nodeApi, minerApi),
	}
}

type CachedFullNode struct {
	APISecret *dtypes.APIAlg
	nodeApi   api.FullNode
	minerApi  api.StorageMiner
	cache     *cache.Cache
	wrapper   *apiwrapper.LotusAPIWrapper
}

func (c *CachedFullNode) AuthVerify(_ context.Context, token string) ([]auth.Permission, error) {
	return AuthVerify(token, c.APISecret)
}

func (c *CachedFullNode) MinerProvingInfo(ctx context.Context, miner address.Address) (*apitypes.ProvingInfo, error) {
	k := fmt.Sprintf("MinerProvingInfo")
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(*apitypes.ProvingInfo), nil
	}
	info, err := c.wrapper.MinerProvingInfo(ctx, miner)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(k, info)
	return info, nil
}

func (c *CachedFullNode) SectorsStatus(ctx context.Context, sid abi.SectorNumber, showOnChainInfo bool) (api.SectorInfo, error) {
	k := fmt.Sprintf("SectorsStatus%d", sid)
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(api.SectorInfo), nil
	}
	info, err := c.minerApi.SectorsStatus(ctx, sid, showOnChainInfo)
	if err != nil {
		return api.SectorInfo{}, err
	}
	c.cache.SetDefault(k, info)
	return info, nil
}

func (c *CachedFullNode) WorkerStats(ctx context.Context) (map[uuid.UUID]storiface.WorkerStats, error) {
	k := fmt.Sprintf("WorkerStats")
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(map[uuid.UUID]storiface.WorkerStats), nil
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

func (c *CachedFullNode) WorkerJobs(ctx context.Context) (map[uuid.UUID][]storiface.WorkerJob, error) {
	k := fmt.Sprintf("WorkerJobs")
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(map[uuid.UUID][]storiface.WorkerJob), nil
	}
	info, err := c.minerApi.WorkerJobs(ctx)
	if err != nil {
		return nil, err
	}
	c.cache.Set(k, info, time.Second)
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
	return c.wrapper.MinerAssetInfo(ctx, mAddr)
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
