package apitypes

import (
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
	ProvingInfo      *ProvingInfo       `json:"proving_info"`
	MinerSectorsInfo *MinerSectorsInfo  `json:"miner_sectors_info"`
	WorkerTaskState  []*WorkerTaskState `json:"worker_task_state"`
	ClusterAssetInfo *ClusterAssetInfo  `json:"cluster_asset_info"`
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
