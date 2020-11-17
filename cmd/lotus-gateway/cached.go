package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apibstore"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
	"github.com/filecoin-project/lotus/lib/blockstore"
	"github.com/filecoin-project/lotus/lib/bufbstore"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/filecoin-project/lotus/storage"
	"github.com/google/uuid"
	"github.com/guoxiaopeng875/lotus-adapter/api/apitypes"
	"github.com/hako/durafmt"
	cbor "github.com/ipfs/go-ipld-cbor"
	"github.com/patrickmn/go-cache"
	"golang.org/x/xerrors"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
)

func NewCachedFullNode(nodeApi api.FullNode, minerApi api.StorageMiner, cache *cache.Cache, secret *dtypes.APIAlg) *CachedFullNode {
	return &CachedFullNode{nodeApi: nodeApi, minerApi: minerApi, cache: cache, APISecret: secret}
}

type CachedFullNode struct {
	APISecret *dtypes.APIAlg
	nodeApi   api.FullNode
	minerApi  api.StorageMiner
	cache     *cache.Cache
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
	info, err := c.minerProvingInfo(ctx, miner)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(k, info)
	return info, nil
}

func (c *CachedFullNode) minerProvingInfo(ctx context.Context, mAddr address.Address) (*apitypes.ProvingInfo, error) {
	node := c.nodeApi
	head, err := node.ChainHead(ctx)
	if err != nil {
		return nil, xerrors.Errorf("getting chain head: %w", err)
	}

	mact, err := node.StateGetActor(ctx, mAddr, head.Key())
	if err != nil {
		return nil, err
	}

	stor := store.ActorStore(ctx, apibstore.NewAPIBlockstore(node))

	mas, err := miner.Load(stor, mact)
	if err != nil {
		return nil, err
	}

	cd, err := node.StateMinerProvingDeadline(ctx, mAddr, head.Key())
	if err != nil {
		return nil, xerrors.Errorf("getting miner info: %w", err)
	}

	proving := uint64(0)
	faults := uint64(0)
	recovering := uint64(0)
	curDeadlineSectors := uint64(0)

	if err := mas.ForEachDeadline(func(dlIdx uint64, dl miner.Deadline) error {
		return dl.ForEachPartition(func(partIdx uint64, part miner.Partition) error {
			if bf, err := part.LiveSectors(); err != nil {
				return err
			} else if count, err := bf.Count(); err != nil {
				return err
			} else {
				proving += count
				if dlIdx == cd.Index {
					curDeadlineSectors += count
				}
			}

			if bf, err := part.FaultySectors(); err != nil {
				return err
			} else if count, err := bf.Count(); err != nil {
				return err
			} else {
				faults += count
			}

			if bf, err := part.RecoveringSectors(); err != nil {
				return err
			} else if count, err := bf.Count(); err != nil {
				return err
			} else {
				recovering += count
			}

			return nil
		})
	}); err != nil {
		return nil, xerrors.Errorf("walking miner deadlines and partitions: %w", err)
	}

	var faultPerc float64
	if proving > 0 {
		faultPerc = float64(faults*10000/proving) / 100
	}

	return &apitypes.ProvingInfo{
		CurrentEpoch:          cd.CurrentEpoch,
		ProvingPeriodBoundary: cd.PeriodStart % cd.WPoStProvingPeriod,
		ProvingPeriodStart:    cli.EpochTime(cd.CurrentEpoch, cd.PeriodStart),
		NextPeriodStart:       cli.EpochTime(cd.CurrentEpoch, cd.PeriodStart+cd.WPoStProvingPeriod),
		Faults:                fmt.Sprintf("%d (%.2f%%)", faults, faultPerc),
		Recovering:            recovering,
		DeadlineIndex:         cd.Index,
		DeadlineSectors:       curDeadlineSectors,
		DeadlineOpen:          cli.EpochTime(cd.CurrentEpoch, cd.Open),
		DeadlineClose:         cli.EpochTime(cd.CurrentEpoch, cd.Close),
		DeadlineElapsed:       durafmt.Parse(time.Second * time.Duration(int64(build.BlockDelaySecs)*int64(cd.Close-cd.Open))).LimitFirstN(2).String(),
		DeadlineChallenge:     cli.EpochTime(cd.CurrentEpoch, cd.Challenge),
		DeadlineFaultCutoff:   cli.EpochTime(cd.CurrentEpoch, cd.FaultCutoff),
	}, nil
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
	wBls, err := c.nodeApi.WalletBalance(ctx, mi.Worker)
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
		WorkerBalance:            wBls,
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
