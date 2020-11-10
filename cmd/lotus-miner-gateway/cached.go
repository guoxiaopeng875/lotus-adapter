package main

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/piecestore"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/extern/sector-storage/fsutil"
	"github.com/filecoin-project/lotus/extern/sector-storage/stores"
	"github.com/filecoin-project/lotus/extern/sector-storage/storiface"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/patrickmn/go-cache"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type CachedStorageMiner struct {
	under api.StorageMiner
	cache *cache.Cache
}

func NewCachedStorageMiner(under api.StorageMiner, cache *cache.Cache) *CachedStorageMiner {
	return &CachedStorageMiner{under: under, cache: cache}
}

func (c CachedStorageMiner) AuthVerify(ctx context.Context, token string) ([]auth.Permission, error) {
	return nil, nil
}

func (c CachedStorageMiner) AuthNew(ctx context.Context, perms []auth.Permission) ([]byte, error) {
	return nil, nil
}

func (c CachedStorageMiner) NetConnectedness(ctx context.Context, id peer.ID) (network.Connectedness, error) {
	return 0, nil
}

func (c CachedStorageMiner) NetPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	return nil, nil
}

func (c CachedStorageMiner) NetConnect(ctx context.Context, info peer.AddrInfo) error {
	return nil
}

func (c CachedStorageMiner) NetAddrsListen(ctx context.Context) (peer.AddrInfo, error) {
	return peer.AddrInfo{}, nil
}

func (c CachedStorageMiner) NetDisconnect(ctx context.Context, id peer.ID) error {
	return nil
}

func (c CachedStorageMiner) NetFindPeer(ctx context.Context, id peer.ID) (peer.AddrInfo, error) {
	return peer.AddrInfo{}, nil
}

func (c CachedStorageMiner) NetPubsubScores(ctx context.Context) ([]api.PubsubScore, error) {
	return nil, nil
}

func (c CachedStorageMiner) NetAutoNatStatus(ctx context.Context) (api.NatInfo, error) {
	return api.NatInfo{}, nil
}

func (c CachedStorageMiner) NetAgentVersion(ctx context.Context, p peer.ID) (string, error) {
	return "", nil
}

func (c CachedStorageMiner) NetBandwidthStats(ctx context.Context) (metrics.Stats, error) {
	return metrics.Stats{}, nil
}

func (c CachedStorageMiner) NetBandwidthStatsByPeer(ctx context.Context) (map[string]metrics.Stats, error) {
	return nil, nil
}

func (c CachedStorageMiner) NetBandwidthStatsByProtocol(ctx context.Context) (map[protocol.ID]metrics.Stats, error) {
	return nil, nil
}

func (c CachedStorageMiner) ID(ctx context.Context) (peer.ID, error) {
	return "", nil
}

func (c CachedStorageMiner) Version(ctx context.Context) (api.Version, error) {
	return c.under.Version(ctx)
}

func (c CachedStorageMiner) LogList(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (c CachedStorageMiner) LogSetLevel(ctx context.Context, s string, s2 string) error {
	return nil
}

func (c CachedStorageMiner) Shutdown(ctx context.Context) error {
	return nil
}

func (c CachedStorageMiner) Closing(ctx context.Context) (<-chan struct{}, error) {
	return nil, nil
}

func (c CachedStorageMiner) ActorAddress(ctx context.Context) (address.Address, error) {
	key := "ActorAddress"
	cachedData, exist := c.cache.Get(key)
	if exist {
		return cachedData.(address.Address), nil
	}
	addr, err := c.under.ActorAddress(ctx)
	if err != nil {
		return address.Address{}, err
	}
	c.cache.SetDefault(key, addr)
	return addr, nil
}

func (c CachedStorageMiner) ActorSectorSize(ctx context.Context, address address.Address) (abi.SectorSize, error) {
	return 0, nil
}

func (c CachedStorageMiner) MiningBase(ctx context.Context) (*types.TipSet, error) {
	return nil, nil
}

func (c CachedStorageMiner) PledgeSector(ctx context.Context) error {
	return nil
}

func (c CachedStorageMiner) SectorsStatus(ctx context.Context, sid abi.SectorNumber, showOnChainInfo bool) (api.SectorInfo, error) {
	key := fmt.Sprintf("SectorsStatus:%d:%v", sid, showOnChainInfo)
	cachedData, exist := c.cache.Get(key)
	if exist {
		return cachedData.(api.SectorInfo), nil
	}
	si, err := c.under.SectorsStatus(ctx, sid, showOnChainInfo)
	if err != nil {
		return api.SectorInfo{}, err
	}
	c.cache.SetDefault(key, si)
	return si, nil
}

func (c CachedStorageMiner) SectorsList(ctx context.Context) ([]abi.SectorNumber, error) {
	key := "SectorsList"
	cachedData, exist := c.cache.Get(key)
	if exist {
		return cachedData.([]abi.SectorNumber), nil
	}
	sl, err := c.under.SectorsList(ctx)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(key, sl)
	return sl, nil
}

func (c CachedStorageMiner) SectorsRefs(ctx context.Context) (map[string][]api.SealedRef, error) {
	return nil, nil
}

func (c CachedStorageMiner) SectorStartSealing(ctx context.Context, number abi.SectorNumber) error {
	return nil
}

func (c CachedStorageMiner) SectorSetSealDelay(ctx context.Context, duration time.Duration) error {
	return nil
}

func (c CachedStorageMiner) SectorGetSealDelay(ctx context.Context) (time.Duration, error) {
	return 0, nil
}

func (c CachedStorageMiner) SectorSetExpectedSealDuration(ctx context.Context, duration time.Duration) error {
	return nil
}

func (c CachedStorageMiner) SectorGetExpectedSealDuration(ctx context.Context) (time.Duration, error) {
	return 0, nil
}

func (c CachedStorageMiner) SectorsUpdate(ctx context.Context, number abi.SectorNumber, state api.SectorState) error {
	return nil
}

func (c CachedStorageMiner) SectorRemove(ctx context.Context, number abi.SectorNumber) error {
	return nil
}

func (c CachedStorageMiner) SectorMarkForUpgrade(ctx context.Context, id abi.SectorNumber) error {
	return nil
}

func (c CachedStorageMiner) StorageList(ctx context.Context) (map[stores.ID][]stores.Decl, error) {
	return nil, nil
}

func (c CachedStorageMiner) StorageLocal(ctx context.Context) (map[stores.ID]string, error) {
	return nil, nil
}

func (c CachedStorageMiner) StorageStat(ctx context.Context, id stores.ID) (fsutil.FsStat, error) {
	return fsutil.FsStat{}, nil
}

func (c CachedStorageMiner) WorkerConnect(ctx context.Context, s string) error {
	return nil
}

func (c CachedStorageMiner) WorkerStats(ctx context.Context) (map[uint64]storiface.WorkerStats, error) {
	key := "WorkerStats"
	cachedData, exist := c.cache.Get(key)
	if exist {
		return cachedData.(map[uint64]storiface.WorkerStats), nil
	}
	ws, err := c.under.WorkerStats(ctx)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(key, ws)
	return ws, nil
}

func (c CachedStorageMiner) WorkerJobs(ctx context.Context) (map[uint64][]storiface.WorkerJob, error) {
	key := "WorkerJobs"
	cachedData, exist := c.cache.Get(key)
	if exist {
		return cachedData.(map[uint64][]storiface.WorkerJob), nil
	}
	wj, err := c.under.WorkerJobs(ctx)
	if err != nil {
		return nil, err
	}
	c.cache.SetDefault(key, wj)
	return wj, nil
}

func (c CachedStorageMiner) SealingSchedDiag(ctx context.Context) (interface{}, error) {
	return nil, nil
}

func (c CachedStorageMiner) StorageAttach(ctx context.Context, info stores.StorageInfo, stat fsutil.FsStat) error {
	return nil
}

func (c CachedStorageMiner) StorageInfo(ctx context.Context, id stores.ID) (stores.StorageInfo, error) {
	return stores.StorageInfo{}, nil
}

func (c CachedStorageMiner) StorageReportHealth(ctx context.Context, id stores.ID, report stores.HealthReport) error {
	return nil
}

func (c CachedStorageMiner) StorageDeclareSector(ctx context.Context, storageID stores.ID, s abi.SectorID, ft stores.SectorFileType, primary bool) error {
	return nil
}

func (c CachedStorageMiner) StorageDropSector(ctx context.Context, storageID stores.ID, s abi.SectorID, ft stores.SectorFileType) error {
	return nil
}

func (c CachedStorageMiner) StorageFindSector(ctx context.Context, sector abi.SectorID, ft stores.SectorFileType, ssize abi.SectorSize, allowFetch bool) ([]stores.SectorStorageInfo, error) {
	return nil, nil
}

func (c CachedStorageMiner) StorageBestAlloc(ctx context.Context, allocate stores.SectorFileType, ssize abi.SectorSize, pathType stores.PathType) ([]stores.StorageInfo, error) {
	return nil, nil
}

func (c CachedStorageMiner) StorageLock(ctx context.Context, sector abi.SectorID, read stores.SectorFileType, write stores.SectorFileType) error {
	return nil
}

func (c CachedStorageMiner) StorageTryLock(ctx context.Context, sector abi.SectorID, read stores.SectorFileType, write stores.SectorFileType) (bool, error) {
	return false, nil
}

func (c CachedStorageMiner) MarketImportDealData(ctx context.Context, propcid cid.Cid, path string) error {
	return nil
}

func (c CachedStorageMiner) MarketListDeals(ctx context.Context) ([]api.MarketDeal, error) {
	return nil, nil
}

func (c CachedStorageMiner) MarketListRetrievalDeals(ctx context.Context) ([]retrievalmarket.ProviderDealState, error) {
	return nil, nil
}

func (c CachedStorageMiner) MarketGetDealUpdates(ctx context.Context) (<-chan storagemarket.MinerDeal, error) {
	return nil, nil
}

func (c CachedStorageMiner) MarketListIncompleteDeals(ctx context.Context) ([]storagemarket.MinerDeal, error) {
	return nil, nil
}

func (c CachedStorageMiner) MarketSetAsk(ctx context.Context, price types.BigInt, verifiedPrice types.BigInt, duration abi.ChainEpoch, minPieceSize abi.PaddedPieceSize, maxPieceSize abi.PaddedPieceSize) error {
	return nil
}

func (c CachedStorageMiner) MarketGetAsk(ctx context.Context) (*storagemarket.SignedStorageAsk, error) {
	return nil, nil
}

func (c CachedStorageMiner) MarketSetRetrievalAsk(ctx context.Context, rask *retrievalmarket.Ask) error {
	return nil
}

func (c CachedStorageMiner) MarketGetRetrievalAsk(ctx context.Context) (*retrievalmarket.Ask, error) {
	return nil, nil
}

func (c CachedStorageMiner) MarketListDataTransfers(ctx context.Context) ([]api.DataTransferChannel, error) {
	return nil, nil
}

func (c CachedStorageMiner) MarketDataTransferUpdates(ctx context.Context) (<-chan api.DataTransferChannel, error) {
	return nil, nil
}

func (c CachedStorageMiner) DealsImportData(ctx context.Context, dealPropCid cid.Cid, file string) error {
	return nil
}

func (c CachedStorageMiner) DealsList(ctx context.Context) ([]api.MarketDeal, error) {
	return nil, nil
}

func (c CachedStorageMiner) DealsConsiderOnlineStorageDeals(ctx context.Context) (bool, error) {
	return false, nil
}

func (c CachedStorageMiner) DealsSetConsiderOnlineStorageDeals(ctx context.Context, b bool) error {
	return nil
}

func (c CachedStorageMiner) DealsConsiderOnlineRetrievalDeals(ctx context.Context) (bool, error) {
	return false, nil
}

func (c CachedStorageMiner) DealsSetConsiderOnlineRetrievalDeals(ctx context.Context, b bool) error {
	return nil
}

func (c CachedStorageMiner) DealsPieceCidBlocklist(ctx context.Context) ([]cid.Cid, error) {
	return nil, nil
}

func (c CachedStorageMiner) DealsSetPieceCidBlocklist(ctx context.Context, cids []cid.Cid) error {
	return nil
}

func (c CachedStorageMiner) DealsConsiderOfflineStorageDeals(ctx context.Context) (bool, error) {
	return false, nil
}

func (c CachedStorageMiner) DealsSetConsiderOfflineStorageDeals(ctx context.Context, b bool) error {
	return nil
}

func (c CachedStorageMiner) DealsConsiderOfflineRetrievalDeals(ctx context.Context) (bool, error) {
	return false, nil
}

func (c CachedStorageMiner) DealsSetConsiderOfflineRetrievalDeals(ctx context.Context, b bool) error {
	return nil
}

func (c CachedStorageMiner) StorageAddLocal(ctx context.Context, path string) error {
	return nil
}

func (c CachedStorageMiner) PiecesListPieces(ctx context.Context) ([]cid.Cid, error) {
	return nil, nil
}

func (c CachedStorageMiner) PiecesListCidInfos(ctx context.Context) ([]cid.Cid, error) {
	return nil, nil
}

func (c CachedStorageMiner) PiecesGetPieceInfo(ctx context.Context, pieceCid cid.Cid) (*piecestore.PieceInfo, error) {
	return nil, nil
}

func (c CachedStorageMiner) PiecesGetCIDInfo(ctx context.Context, payloadCid cid.Cid) (*piecestore.CIDInfo, error) {
	return nil, nil
}

func (c CachedStorageMiner) CreateBackup(ctx context.Context, fpath string) error {
	return nil
}
