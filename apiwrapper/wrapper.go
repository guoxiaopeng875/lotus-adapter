package apiwrapper

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apibstore"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/actors/adt"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/extern/sector-storage/fsutil"
	"github.com/filecoin-project/lotus/extern/sector-storage/stores"
	"github.com/filecoin-project/lotus/lib/blockstore"
	"github.com/filecoin-project/lotus/lib/bufbstore"
	"github.com/filecoin-project/lotus/storage"
	"github.com/guoxiaopeng875/lotus-adapter/api/apitypes"
	"github.com/hako/durafmt"
	cbor "github.com/ipfs/go-ipld-cbor"
	"golang.org/x/xerrors"
	"sort"
	"strings"
	"time"
)

type fsInfo struct {
	stores.ID
	sectors []stores.Decl
	stat    fsutil.FsStat
}

type LotusAPIWrapper struct {
	api.FullNode
	api.StorageMiner
}

func NewLotusAPIWrapper(fullNode api.FullNode, storageMiner api.StorageMiner) *LotusAPIWrapper {
	return &LotusAPIWrapper{FullNode: fullNode, StorageMiner: storageMiner}
}

func (c *LotusAPIWrapper) MinerProvingInfo(ctx context.Context, mAddr address.Address) (*apitypes.ProvingInfo, error) {
	node := c.FullNode
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

func (c *LotusAPIWrapper) MinerAssetInfo(ctx context.Context, mAddr address.Address) (*apitypes.ClusterAssetInfo, error) {
	nodeApi := c.FullNode
	mi, err := nodeApi.StateMinerInfo(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}
	mAct, err := nodeApi.StateGetActor(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}
	power, err := nodeApi.StateMinerPower(ctx, mAddr, types.EmptyTSK)
	if err != nil {
		return nil, err
	}
	tbs := bufbstore.NewTieredBstore(apibstore.NewAPIBlockstore(nodeApi), blockstore.NewTemporary())
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
	postAddr, err := storage.AddressFor(ctx, nodeApi, mi, storage.PoStAddr, types.FromFil(1))
	if err != nil {
		return nil, xerrors.Errorf("getting address for post: %w", err)
	}
	postBls, err := nodeApi.WalletBalance(ctx, postAddr)
	if err != nil {
		return nil, err
	}
	wBls, err := nodeApi.WalletBalance(ctx, mi.Worker)
	if err != nil {
		return nil, err
	}
	ownerBls, err := nodeApi.WalletBalance(ctx, mi.Owner)
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
		QualityAdjPower:          power.MinerPower.QualityAdjPower,
		OwnerBalance:             ownerBls,
	}, nil
}

// 扇区信息
func (c *LotusAPIWrapper) SectorsInfo() (*apitypes.MinerSectorsInfo, error) {
	ctx := context.Background()
	node := c.StorageMiner
	sectors, err := node.SectorsList(ctx)
	if err != nil {
		return nil, err
	}
	msi := &apitypes.MinerSectorsInfo{
		TotalSectors: len(sectors),
	}
	for _, sid := range sectors {
		si, err := node.SectorsStatus(ctx, sid, false)
		if err != nil {
			return nil, err
		}
		switch si.State {
		// TODO 定义常量
		case "Proving":
			msi.Proving++
		default:
			continue
		}
	}
	return msi, nil
}

// worker任务信息
func (c *LotusAPIWrapper) WorkerTaskInfo() ([]*apitypes.WorkerTaskState, error) {
	minerAPI := c.StorageMiner
	ctx := context.Background()

	var wtStates []*apitypes.WorkerTaskState
	jobs, err := minerAPI.WorkerJobs(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := minerAPI.WorkerStats(ctx)
	if err != nil {
		return nil, err
	}
	for workerID, st := range stats {
		wts := &apitypes.WorkerTaskState{
			ID:           workerID.String(),
			Hostname:     st.Info.Hostname,
			Enable:       st.Enabled,
			SectorStates: nil,
		}
		workerJobs, ok := jobs[workerID]
		if !ok {
			wtStates = append(wtStates, wts)
			continue
		}
		for _, job := range workerJobs {
			ss := &apitypes.SectorState{
				SectorNum: uint64(job.Sector.Number),
				Start:     job.Start,
				RunWait:   job.RunWait,
			}
			ss.Task = strings.TrimSpace(job.Task.Short())
			wts.SectorStates = append(wts.SectorStates, ss)
		}
		wtStates = append(wtStates, wts)
	}
	return wtStates, nil
}

func (c *LotusAPIWrapper) GetStorageInfo() ([]*apitypes.StorageInfo, error) {
	minerAPI := c.StorageMiner
	ctx := context.Background()

	st, err := minerAPI.StorageList(ctx)
	if err != nil {
		return nil, err
	}

	sorted := make([]*fsInfo, 0, len(st))
	for id, decls := range st {
		st, err := minerAPI.StorageStat(ctx, id)
		if err != nil {
			sorted = append(sorted, &fsInfo{ID: id, sectors: decls})
			continue
		}

		sorted = append(sorted, &fsInfo{id, decls, st})
	}

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].stat.Capacity != sorted[j].stat.Capacity {
			return sorted[i].stat.Capacity > sorted[j].stat.Capacity
		}
		return sorted[i].ID < sorted[j].ID
	})

	storageInfos := make([]*apitypes.StorageInfo, len(sorted))
	for i, fs := range sorted {
		storageInfos[i] = &apitypes.StorageInfo{
			ID:        string(fs.ID),
			Sectors:   make([]*apitypes.Decl, len(fs.sectors)),
			Capacity:  fs.stat.Capacity,
			Available: fs.stat.Available,
			Reserved:  fs.stat.Reserved,
		}

		for j, sector := range fs.sectors {
			storageInfos[i].Sectors[j] = &apitypes.Decl{
				Miner:          sector.Miner.String(),
				SectorNumber:   sector.Number.String(),
				SectorFileType: sector.SectorFileType.String(),
			}
		}
	}

	return storageInfos, nil
}
