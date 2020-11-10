package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/lotus/api"
	"github.com/patrickmn/go-cache"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
)

type CachedFullNode struct {
	under api.FullNode
	cache *cache.Cache
}

func NewCachedFullNode(under api.FullNode, cache *cache.Cache) *CachedFullNode {
	return &CachedFullNode{under: under, cache: cache}
}

func (c *CachedFullNode) WalletBalance(ctx context.Context, address address.Address) (types.BigInt, error) {
	k := fmt.Sprintf("WalletBalance:%s", address.String())
	cachedData, exist := c.cache.Get(k)
	if exist {
		return cachedData.(types.BigInt), nil
	}
	wb, err := c.under.WalletBalance(ctx, address)
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
	act, err := c.under.StateGetActor(ctx, actor, tsk)
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
	mi, err := c.under.StateMinerInfo(ctx, address, key)
	if err != nil {
		return miner.MinerInfo{}, err
	}
	c.cache.SetDefault(k, mi)
	return mi, nil
}
