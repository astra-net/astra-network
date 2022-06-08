package astra

import (
	"context"
	"fmt"
	"math/big"

	"github.com/astra-net/astra-network/block"
	"github.com/astra-net/astra-network/core"
	"github.com/astra-net/astra-network/core/rawdb"
	"github.com/astra-net/astra-network/core/state"
	"github.com/astra-net/astra-network/core/types"
	"github.com/astra-net/astra-network/crypto/bls"
	internal_bls "github.com/astra-net/astra-network/crypto/bls"
	"github.com/astra-net/astra-network/eth/rpc"
	"github.com/astra-net/astra-network/internal/params"
	"github.com/astra-net/astra-network/internal/utils"
	"github.com/astra-net/astra-network/shard"
	"github.com/astra-net/astra-network/staking/availability"
	stakingReward "github.com/astra-net/astra-network/staking/reward"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
)

// ChainConfig ...
func (astra *Astra) ChainConfig() *params.ChainConfig {
	return astra.BlockChain.Config()
}

// GetShardState ...
func (astra *Astra) GetShardState() (*shard.State, error) {
	return astra.BlockChain.ReadShardState(astra.BlockChain.CurrentHeader().Epoch())
}

// GetBlockSigners ..
func (astra *Astra) GetBlockSigners(
	ctx context.Context, blockNum rpc.BlockNumber,
) (shard.SlotList, *internal_bls.Mask, error) {
	blk, err := astra.BlockByNumber(ctx, blockNum)
	if err != nil {
		return nil, nil, err
	}
	blockWithSigners, err := astra.BlockByNumber(ctx, blockNum+1)
	if err != nil {
		return nil, nil, err
	}
	if blockWithSigners == nil {
		return nil, nil, fmt.Errorf("block number %v not found", blockNum+1)
	}
	committee, err := astra.GetValidators(blk.Epoch())
	if err != nil {
		return nil, nil, err
	}
	pubKeys := make([]internal_bls.PublicKeyWrapper, len(committee.Slots))
	for i, validator := range committee.Slots {
		key, err := bls.BytesToBLSPublicKey(validator.BLSPublicKey[:])
		if err != nil {
			return nil, nil, err
		}
		pubKeys[i] = internal_bls.PublicKeyWrapper{
			Bytes:  validator.BLSPublicKey,
			Object: key,
		}
	}
	mask, err := internal_bls.NewMask(pubKeys, nil)
	if err != nil {
		return nil, nil, err
	}
	err = mask.SetMask(blockWithSigners.Header().LastCommitBitmap())
	if err != nil {
		return nil, nil, err
	}
	return committee.Slots, mask, nil
}

// DetailedBlockSignerInfo contains all of the block singing information
type DetailedBlockSignerInfo struct {
	// Signers are all the signers for the block
	Signers shard.SlotList
	// Committee when the block was signed.
	Committee shard.SlotList
	BlockHash common.Hash
}

// GetDetailedBlockSignerInfo fetches the block signer information for any non-genesis block
func (astra *Astra) GetDetailedBlockSignerInfo(
	ctx context.Context, blk *types.Block,
) (*DetailedBlockSignerInfo, error) {
	parentBlk, err := astra.BlockByNumber(ctx, rpc.BlockNumber(blk.NumberU64()-1))
	if err != nil {
		return nil, err
	}
	parentShardState, err := astra.BlockChain.ReadShardState(parentBlk.Epoch())
	if err != nil {
		return nil, err
	}
	committee, signers, _, err := availability.BallotResult(
		parentBlk.Header(), blk.Header(), parentShardState, blk.ShardID(),
	)
	return &DetailedBlockSignerInfo{
		Signers:   signers,
		Committee: committee,
		BlockHash: blk.Hash(),
	}, nil
}

// PreStakingBlockRewards are the rewards for a block in the pre-staking era (epoch < staking epoch).
type PreStakingBlockRewards map[common.Address]*big.Int

// GetPreStakingBlockRewards for the given block number.
// Calculated rewards are done exactly like chain.AccumulateRewardsAndCountSigs.
func (astra *Astra) GetPreStakingBlockRewards(
	ctx context.Context, blk *types.Block,
) (PreStakingBlockRewards, error) {
	if astra.IsStakingEpoch(blk.Epoch()) {
		return nil, fmt.Errorf("block %v is in staking era", blk.Number())
	}

	if cachedReward, ok := astra.preStakingBlockRewardsCache.Get(blk.Hash()); ok {
		return cachedReward.(PreStakingBlockRewards), nil
	}
	rewards := PreStakingBlockRewards{}

	sigInfo, err := astra.GetDetailedBlockSignerInfo(ctx, blk)
	if err != nil {
		return nil, err
	}
	last := big.NewInt(0)
	count := big.NewInt(int64(len(sigInfo.Signers)))
	for i, slot := range sigInfo.Signers {
		rewardsForThisAddr, ok := rewards[slot.EcdsaAddress]
		if !ok {
			rewardsForThisAddr = big.NewInt(0)
		}
		cur := big.NewInt(0)
		cur.Mul(stakingReward.PreStakedBlocks, big.NewInt(int64(i+1))).Div(cur, count)
		reward := big.NewInt(0).Sub(cur, last)
		rewards[slot.EcdsaAddress] = new(big.Int).Add(reward, rewardsForThisAddr)
		last = cur
	}

	// Report tx fees of the coinbase (== leader)
	receipts, err := astra.GetReceipts(ctx, blk.Hash())
	if err != nil {
		return nil, err
	}
	txFees := big.NewInt(0)
	for _, tx := range blk.Transactions() {
		txnHash := tx.HashByType()
		dbTx, _, _, receiptIndex := rawdb.ReadTransaction(astra.ChainDb(), txnHash)
		if dbTx == nil {
			return nil, fmt.Errorf("could not find receipt for tx: %v", txnHash.String())
		}
		if len(receipts) <= int(receiptIndex) {
			return nil, fmt.Errorf("invalid receipt indext %v (>= num receipts: %v) for tx: %v",
				receiptIndex, len(receipts), txnHash.String())
		}
		txFee := new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(receipts[receiptIndex].GasUsed)))
		txFees = new(big.Int).Add(txFee, txFees)
	}

	if amt, ok := rewards[blk.Header().Coinbase()]; ok {
		rewards[blk.Header().Coinbase()] = new(big.Int).Add(amt, txFees)
	} else {
		rewards[blk.Header().Coinbase()] = txFees
	}

	astra.preStakingBlockRewardsCache.Add(blk.Hash(), rewards)
	return rewards, nil
}

// GetLatestChainHeaders ..
func (astra *Astra) GetLatestChainHeaders() *block.HeaderPair {
	return &block.HeaderPair{
		BeaconHeader: astra.BeaconChain.CurrentHeader(),
		ShardHeader:  astra.BlockChain.CurrentHeader(),
	}
}

// GetLastCrossLinks ..
func (astra *Astra) GetLastCrossLinks() ([]*types.CrossLink, error) {
	crossLinks := []*types.CrossLink{}
	for i := uint32(1); i < shard.Schedule.InstanceForEpoch(astra.CurrentBlock().Epoch()).NumShards(); i++ {
		link, err := astra.BlockChain.ReadShardLastCrossLink(i)
		if err != nil {
			return nil, err
		}
		crossLinks = append(crossLinks, link)
	}

	return crossLinks, nil
}

// CurrentBlock ...
func (astra *Astra) CurrentBlock() *types.Block {
	return types.NewBlockWithHeader(astra.BlockChain.CurrentHeader())
}

// GetBlock ...
func (astra *Astra) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return astra.BlockChain.GetBlockByHash(hash), nil
}

// GetCurrentBadBlocks ..
func (astra *Astra) GetCurrentBadBlocks() []core.BadBlock {
	return astra.BlockChain.BadBlocks()
}

func (astra *Astra) BlockByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*types.Block, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return astra.BlockByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header := astra.BlockChain.GetHeaderByHash(hash)
		if header == nil {
			return nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && astra.BlockChain.GetCanonicalHash(header.Number().Uint64()) != hash {
			return nil, errors.New("hash is not currently canonical")
		}
		block := astra.BlockChain.GetBlock(hash, header.Number().Uint64())
		if block == nil {
			return nil, errors.New("header found, but block body is missing")
		}
		return block, nil
	}
	return nil, errors.New("invalid arguments; neither block nor hash specified")
}

// GetBalance returns balance of an given address.
func (astra *Astra) GetBalance(ctx context.Context, address common.Address, blockNrOrHash rpc.BlockNumberOrHash) (*big.Int, error) {
	s, _, err := astra.StateAndHeaderByNumberOrHash(ctx, blockNrOrHash)
	if s == nil || err != nil {
		return nil, err
	}
	return s.GetBalance(address), s.Error()
}

// BlockByNumber ...
func (astra *Astra) BlockByNumber(ctx context.Context, blockNum rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNum == rpc.PendingBlockNumber {
		return nil, errors.New("not implemented")
	}
	// Otherwise resolve and return the block
	if blockNum == rpc.LatestBlockNumber {
		return astra.BlockChain.CurrentBlock(), nil
	}
	return astra.BlockChain.GetBlockByNumber(uint64(blockNum)), nil
}

// HeaderByNumber ...
func (astra *Astra) HeaderByNumber(ctx context.Context, blockNum rpc.BlockNumber) (*block.Header, error) {
	// Pending block is only known by the miner
	if blockNum == rpc.PendingBlockNumber {
		return nil, errors.New("not implemented")
	}
	// Otherwise resolve and return the block
	if blockNum == rpc.LatestBlockNumber {
		return astra.BlockChain.CurrentBlock().Header(), nil
	}
	return astra.BlockChain.GetHeaderByNumber(uint64(blockNum)), nil
}

// HeaderByHash ...
func (astra *Astra) HeaderByHash(ctx context.Context, blockHash common.Hash) (*block.Header, error) {
	header := astra.BlockChain.GetHeaderByHash(blockHash)
	if header == nil {
		return nil, errors.New("Header is not found")
	}
	return header, nil
}

// StateAndHeaderByNumber ...
func (astra *Astra) StateAndHeaderByNumber(ctx context.Context, blockNum rpc.BlockNumber) (*state.DB, *block.Header, error) {
	// Pending state is only known by the miner
	if blockNum == rpc.PendingBlockNumber {
		return nil, nil, errors.New("not implemented")
	}
	// Otherwise resolve the block number and return its state
	header, err := astra.HeaderByNumber(ctx, blockNum)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := astra.BlockChain.StateAt(header.Root())
	return stateDb, header, err
}

func (astra *Astra) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.DB, *block.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return astra.StateAndHeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		header, err := astra.HeaderByHash(ctx, hash)
		if err != nil {
			return nil, nil, err
		}
		if header == nil {
			return nil, nil, errors.New("header for hash not found")
		}
		if blockNrOrHash.RequireCanonical && astra.BlockChain.GetCanonicalHash(header.Number().Uint64()) != hash {
			return nil, nil, errors.New("hash is not currently canonical")
		}
		stateDb, err := astra.BlockChain.StateAt(header.Root())
		return stateDb, header, err
	}
	return nil, nil, errors.New("invalid arguments; neither block nor hash specified")
}

// GetLeaderAddress returns the one address of the leader, given the coinbaseAddr.
// Note that the coinbaseAddr is overloaded with the BLS pub key hash in staking era.
func (astra *Astra) GetLeaderAddress(coinbaseAddr common.Address, epoch *big.Int) string {
	if astra.IsStakingEpoch(epoch) {
		if leader, exists := astra.leaderCache.Get(coinbaseAddr); exists {
			addr := leader.(common.Address).String()
			return addr
		}
		committee, err := astra.GetValidators(epoch)
		if err != nil {
			return ""
		}
		for _, val := range committee.Slots {
			addr := utils.GetAddressFromBLSPubKeyBytes(val.BLSPublicKey[:])
			astra.leaderCache.Add(addr, val.EcdsaAddress)
			if addr == coinbaseAddr {
				addr := val.EcdsaAddress.String()
				return addr
			}
		}
		return "" // Did not find matching address
	}
	addr := coinbaseAddr.String()
	return addr
}

// Filter related APIs

// GetLogs ...
func (astra *Astra) GetLogs(ctx context.Context, blockHash common.Hash, isEth bool) ([][]*types.Log, error) {
	receipts := astra.BlockChain.GetReceiptsByHash(blockHash)
	if receipts == nil {
		return nil, errors.New("Missing receipts")
	}
	if isEth {
		block := astra.BlockChain.GetBlockByHash(blockHash)
		if block == nil {
			return nil, errors.New("Missing block data")
		}
		txns := block.Transactions()
		for i, _ := range receipts {
			if i < len(txns) {
				ethHash := txns[i].ConvertToEth().Hash()
				receipts[i].TxHash = ethHash
				for j, _ := range receipts[i].Logs {
					// Override log txHash with receipt's
					receipts[i].Logs[j].TxHash = ethHash
				}
			}
		}
	}

	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

// ServiceFilter ...
func (astra *Astra) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	// TODO(dm): implement
}

// SubscribeNewTxsEvent subscribes new tx event.
// TODO: this is not implemented or verified yet for astra.
func (astra *Astra) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	return astra.TxPool.SubscribeNewTxsEvent(ch)
}

// SubscribeChainEvent subscribes chain event.
// TODO: this is not implemented or verified yet for astra.
func (astra *Astra) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return astra.BlockChain.SubscribeChainEvent(ch)
}

// SubscribeChainHeadEvent subcribes chain head event.
// TODO: this is not implemented or verified yet for astra.
func (astra *Astra) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return astra.BlockChain.SubscribeChainHeadEvent(ch)
}

// SubscribeChainSideEvent subcribes chain side event.
// TODO: this is not implemented or verified yet for astra.
func (astra *Astra) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return astra.BlockChain.SubscribeChainSideEvent(ch)
}

// SubscribeRemovedLogsEvent subcribes removed logs event.
// TODO: this is not implemented or verified yet for astra.
func (astra *Astra) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return astra.BlockChain.SubscribeRemovedLogsEvent(ch)
}

// SubscribeLogsEvent subcribes log event.
// TODO: this is not implemented or verified yet for astra.
func (astra *Astra) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return astra.BlockChain.SubscribeLogsEvent(ch)
}
