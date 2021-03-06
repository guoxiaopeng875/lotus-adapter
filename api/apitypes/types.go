package apitypes

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
	"time"
)

// FIL相关都是以attoFIL为单位
// 1FIL = 10e18 attoFIL
type ClusterAssetInfo struct {
	// 矿工号
	MinerID      string  `json:"miner_id"`
	MinerBalance big.Int `json:"miner_balance"`
	// Vesting
	VestingFunds abi.TokenAmount `json:"vesting_funds"`
	// Pledge
	InitialPledgeRequirement abi.TokenAmount `json:"initial_pledge_requirement"`
	// PreCommit
	PreCommitDeposits abi.TokenAmount `json:"pre_commit_deposits"`
	// Available
	AvailableBalance abi.TokenAmount `json:"available_balance"`
	// POST
	PostBalance big.Int `json:"post_balance"`
	// Worker
	WorkerBalance   big.Int          `json:"worker_balance"`
	QualityAdjPower abi.StoragePower `json:"quality_adj_power"`
	//Owner
	OwnerBalance big.Int `json:"owner_balance"`
}

type ProvingInfo struct {
	CurrentEpoch          abi.ChainEpoch `json:"current_epoch"`
	ProvingPeriodBoundary abi.ChainEpoch `json:"proving_period_boundary"`
	ProvingPeriodStart    string         `json:"proving_period_start"`
	NextPeriodStart       string         `json:"next_period_start"`
	Faults                string         `json:"faults"`
	Recovering            uint64         `json:"recovering"`
	DeadlineIndex         uint64         `json:"deadline_index"`
	DeadlineSectors       uint64         `json:"deadline_sectors"`
	DeadlineOpen          string         `json:"deadline_open"`
	DeadlineClose         string         `json:"deadline_close"`
	DeadlineElapsed       string         `json:"deadline_elapsed"`
	DeadlineChallenge     string         `json:"deadline_challenge"`
	DeadlineFaultCutoff   string         `json:"deadline_fault_cutoff"`
}

type MinerSectorsInfo struct {
	TotalSectors int `json:"total_sectors"`
	Proving      int `json:"proving"`
}

type PushedMinerInfo struct {
	MinerID          string             `json:"miner_id"`
	ProvingInfo      *ProvingInfo       `json:"proving_info"`
	MinerSectorsInfo *MinerSectorsInfo  `json:"miner_sectors_info"`
	WorkerTaskState  []*WorkerTaskState `json:"worker_task_state"`
	ClusterAssetInfo *ClusterAssetInfo  `json:"cluster_asset_info"`
	StorageInfo      []*StorageInfo     `json:"storage_info"`
	MessageCount     int                `json:"message_count"`
}

// worker任务状态
type WorkerTaskState struct {
	// workerID
	ID           string         `json:"id"`
	Hostname     string         `json:"hostname"`
	Enable       bool           `json:"enable"`
	SectorStates []*SectorState `json:"sector_states"`
}

type SectorState struct {
	Task      string    `json:"task"`
	SectorNum uint64    `json:"sector_num"`
	Start     time.Time `json:"start"`
	RunWait   int       `json:"run_wait"` // 0 - running, 1+ - assigned
	TaskTime  int64     `json:"task_time,omitempty"`
}

type StorageInfo struct {
	ID        string   `json:"id"`
	Sectors   []*Decl  `json:"sectors"`
	Capacity  int64    `json:"capacity"`
	Available int64    `json:"available"` // Available to use for sector storage
	Reserved  int64    `json:"reserved"`
	URLs      []string `json:"urls"` // TODO: Support non-http transports
	Weight    uint64   `json:"weight"`
	CanSeal   bool     `json:"can_seal"`
	CanStore  bool     `json:"can_store"`
	Local     string   `json:"local"`
}

type Decl struct {
	Miner          string `json:"miner"`
	SectorNumber   string `json:"sector_number"`
	SectorFileType string `json:"sector_file_type"`
}

type Message struct {
	ID         string          `json:"id"`
	Version    uint64          `json:"version"`
	To         address.Address `json:"to"`
	From       address.Address `json:"from"`
	Nonce      uint64          `json:"nonce"`
	Value      abi.TokenAmount `json:"value"`
	GasLimit   int64           `json:"gas_limit"`
	GasFeeCap  abi.TokenAmount `json:"gas_fee_cap"`
	GasPremium abi.TokenAmount `json:"gas_premium"`
	Method     abi.MethodNum   `json:"method"`
	Params     []byte          `json:"params"`
	Type       string          `json:"type"`
	Data       []byte          `json:"data"`
}
