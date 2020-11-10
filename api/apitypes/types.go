package apitypes

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/big"
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
}
