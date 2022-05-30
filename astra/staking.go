package astra

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/astra-net/astra-network/block"
	"github.com/astra-net/astra-network/consensus/quorum"
	"github.com/astra-net/astra-network/core/rawdb"
	"github.com/astra-net/astra-network/core/types"
	"github.com/astra-net/astra-network/eth/rpc"
	"github.com/astra-net/astra-network/internal/chain"
	"github.com/astra-net/astra-network/numeric"
	commonRPC "github.com/astra-net/astra-network/rpc/common"
	"github.com/astra-net/astra-network/shard"
	"github.com/astra-net/astra-network/shard/committee"
	"github.com/astra-net/astra-network/staking/availability"
	"github.com/astra-net/astra-network/staking/effective"
	staking "github.com/astra-net/astra-network/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var (
	zero    = numeric.ZeroDec()
	bigZero = big.NewInt(0)
)

func (astra *Astra) readAndUpdateRawStakes(
	epoch *big.Int,
	decider quorum.Decider,
	comm shard.Committee,
	rawStakes []effective.SlotPurchase,
	validatorSpreads map[common.Address]numeric.Dec,
) []effective.SlotPurchase {
	for i := range comm.Slots {
		slot := comm.Slots[i]
		slotAddr := slot.EcdsaAddress
		slotKey := slot.BLSPublicKey
		spread, ok := validatorSpreads[slotAddr]
		if !ok {
			snapshot, err := astra.BlockChain.ReadValidatorSnapshotAtEpoch(epoch, slotAddr)
			if err != nil {
				continue
			}
			spread = snapshot.RawStakePerSlot()
			validatorSpreads[slotAddr] = spread
		}

		commonRPC.SetRawStake(decider, slotKey, spread)
		// add entry to array for median calculation
		rawStakes = append(rawStakes, effective.SlotPurchase{
			Addr:      slotAddr,
			Key:       slotKey,
			RawStake:  spread,
			EPoSStake: spread,
		})
	}
	return rawStakes
}

func (astra *Astra) getSuperCommittees() (*quorum.Transition, error) {
	nowE := astra.BlockChain.CurrentHeader().Epoch()

	if astra.BlockChain.CurrentHeader().IsLastBlockInEpoch() {
		// current epoch is current header epoch + 1 if the header was last block of prev epoch
		nowE = new(big.Int).Add(nowE, common.Big1)
	}
	thenE := new(big.Int).Sub(nowE, common.Big1)

	var (
		nowCommittee, prevCommittee *shard.State
		err                         error
	)
	nowCommittee, err = astra.BlockChain.ReadShardState(nowE)
	if err != nil {
		return nil, err
	}
	prevCommittee, err = astra.BlockChain.ReadShardState(thenE)
	if err != nil {
		return nil, err
	}

	stakedSlotsNow, stakedSlotsThen :=
		shard.ExternalSlotsAvailableForEpoch(nowE),
		shard.ExternalSlotsAvailableForEpoch(thenE)

	then, now :=
		quorum.NewRegistry(stakedSlotsThen, int(thenE.Int64())),
		quorum.NewRegistry(stakedSlotsNow, int(nowE.Int64()))

	rawStakes := []effective.SlotPurchase{}
	validatorSpreads := map[common.Address]numeric.Dec{}
	for _, comm := range prevCommittee.Shards {
		decider := quorum.NewDecider(quorum.SuperMajorityStake, comm.ShardID)
		// before staking skip computing
		if astra.BlockChain.Config().IsStaking(prevCommittee.Epoch) {
			if _, err := decider.SetVoters(&comm, prevCommittee.Epoch); err != nil {
				return nil, err
			}
		}
		rawStakes = astra.readAndUpdateRawStakes(thenE, decider, comm, rawStakes, validatorSpreads)
		then.Deciders[fmt.Sprintf("shard-%d", comm.ShardID)] = decider
	}
	then.MedianStake = effective.Median(rawStakes)

	rawStakes = []effective.SlotPurchase{}
	validatorSpreads = map[common.Address]numeric.Dec{}
	for _, comm := range nowCommittee.Shards {
		decider := quorum.NewDecider(quorum.SuperMajorityStake, comm.ShardID)
		if _, err := decider.SetVoters(&comm, nowCommittee.Epoch); err != nil {
			return nil, errors.Wrapf(
				err,
				"committee is only available from staking epoch: %v, current epoch: %v",
				astra.BlockChain.Config().StakingEpoch,
				astra.BlockChain.CurrentHeader().Epoch(),
			)
		}
		rawStakes = astra.readAndUpdateRawStakes(nowE, decider, comm, rawStakes, validatorSpreads)
		now.Deciders[fmt.Sprintf("shard-%d", comm.ShardID)] = decider
	}
	now.MedianStake = effective.Median(rawStakes)

	return &quorum.Transition{Previous: then, Current: now}, nil
}

// IsStakingEpoch ...
func (astra *Astra) IsStakingEpoch(epoch *big.Int) bool {
	return astra.BlockChain.Config().IsStaking(epoch)
}

// IsPreStakingEpoch ...
func (astra *Astra) IsPreStakingEpoch(epoch *big.Int) bool {
	return astra.BlockChain.Config().IsPreStaking(epoch)
}

// IsNoEarlyUnlockEpoch ...
func (astra *Astra) IsNoEarlyUnlockEpoch(epoch *big.Int) bool {
	return astra.BlockChain.Config().IsNoEarlyUnlock(epoch)
}

// IsCommitteeSelectionBlock checks if the given block is the committee selection block
func (astra *Astra) IsCommitteeSelectionBlock(header *block.Header) bool {
	return chain.IsCommitteeSelectionBlock(astra.BlockChain, header)
}

// GetDelegationLockingPeriodInEpoch ...
func (astra *Astra) GetDelegationLockingPeriodInEpoch(epoch *big.Int) int {
	return chain.GetLockPeriodInEpoch(astra.BlockChain, epoch)
}

// SendStakingTx adds a staking transaction
func (astra *Astra) SendStakingTx(ctx context.Context, signedStakingTx *staking.StakingTransaction) error {
	stx, _, _, _ := rawdb.ReadStakingTransaction(astra.chainDb, signedStakingTx.Hash())
	if stx == nil {
		return astra.NodeAPI.AddPendingStakingTransaction(signedStakingTx)
	}
	return ErrFinalizedTransaction
}

// GetStakingTransactionsHistory returns list of staking transactions hashes of address.
func (astra *Astra) GetStakingTransactionsHistory(address, txType, order string) ([]common.Hash, error) {
	return astra.NodeAPI.GetStakingTransactionsHistory(address, txType, order)
}

// GetStakingTransactionsCount returns the number of staking transactions of address.
func (astra *Astra) GetStakingTransactionsCount(address, txType string) (uint64, error) {
	return astra.NodeAPI.GetStakingTransactionsCount(address, txType)
}

// GetSuperCommittees ..
func (astra *Astra) GetSuperCommittees() (*quorum.Transition, error) {
	nowE := astra.BlockChain.CurrentHeader().Epoch()
	key := fmt.Sprintf("sc-%s", nowE.String())

	res, err := astra.SingleFlightRequest(
		key, func() (interface{}, error) {
			thenE := new(big.Int).Sub(nowE, common.Big1)
			thenKey := fmt.Sprintf("sc-%s", thenE.String())
			astra.group.Forget(thenKey)
			return astra.getSuperCommittees()
		})
	if err != nil {
		return nil, err
	}
	return res.(*quorum.Transition), err
}

// GetValidators returns validators for a particular epoch.
func (astra *Astra) GetValidators(epoch *big.Int) (*shard.Committee, error) {
	state, err := astra.BlockChain.ReadShardState(epoch)
	if err != nil {
		return nil, err
	}
	for _, cmt := range state.Shards {
		if cmt.ShardID == astra.ShardID {
			return &cmt, nil
		}
	}
	return nil, nil
}

// GetValidatorSelfDelegation returns the amount of staking after applying all delegated stakes
func (astra *Astra) GetValidatorSelfDelegation(addr common.Address) *big.Int {
	wrapper, err := astra.BlockChain.ReadValidatorInformation(addr)
	if err != nil || wrapper == nil {
		return nil
	}
	if len(wrapper.Delegations) == 0 {
		return nil
	}
	return wrapper.Delegations[0].Amount
}

// GetElectedValidatorAddresses returns the address of elected validators for current epoch
func (astra *Astra) GetElectedValidatorAddresses() []common.Address {
	list, _ := astra.BlockChain.ReadShardState(astra.BlockChain.CurrentBlock().Epoch())
	return list.StakedValidators().Addrs
}

// GetAllValidatorAddresses returns the up to date validator candidates for next epoch
func (astra *Astra) GetAllValidatorAddresses() []common.Address {
	return astra.BlockChain.ValidatorCandidates()
}

var (
	epochBlocksMap = map[common.Address]map[uint64]staking.EpochSigningEntry{}
	mapLock        = sync.Mutex{}
)

func (astra *Astra) getEpochSigning(epoch *big.Int, addr common.Address) (staking.EpochSigningEntry, error) {
	entry := staking.EpochSigningEntry{}
	mapLock.Lock()
	defer mapLock.Unlock()
	if validatorMap, ok := epochBlocksMap[addr]; ok {
		if val, ok := validatorMap[epoch.Uint64()]; ok {
			return val, nil
		}
	}

	snapshot, err := astra.BlockChain.ReadValidatorSnapshotAtEpoch(epoch, addr)
	if err != nil {
		return entry, err
	}

	// the signing information is for the previous epoch
	prevEpoch := big.NewInt(0).Sub(epoch, common.Big1)
	entry.Epoch = prevEpoch
	entry.Blocks = snapshot.Validator.Counters

	// subtract previous epoch counters if exists
	prevEpochSnap, err := astra.BlockChain.ReadValidatorSnapshotAtEpoch(prevEpoch, addr)
	if err == nil {
		entry.Blocks.NumBlocksSigned.Sub(
			entry.Blocks.NumBlocksSigned,
			prevEpochSnap.Validator.Counters.NumBlocksSigned,
		)
		entry.Blocks.NumBlocksToSign.Sub(
			entry.Blocks.NumBlocksToSign,
			prevEpochSnap.Validator.Counters.NumBlocksToSign,
		)
	}

	// update map when adding new entry, also remove an entry beyond last 30
	if _, ok := epochBlocksMap[addr]; !ok {
		epochBlocksMap[addr] = map[uint64]staking.EpochSigningEntry{}
	}
	epochBlocksMap[addr][epoch.Uint64()] = entry
	epochMinus30 := big.NewInt(0).Sub(epoch, big.NewInt(staking.SigningHistoryLength))
	delete(epochBlocksMap[addr], epochMinus30.Uint64())

	return entry, nil
}

// GetValidatorInformation returns the information of validator
func (astra *Astra) GetValidatorInformation(
	addr common.Address, block *types.Block,
) (*staking.ValidatorRPCEnhanced, error) {
	bc := astra.BlockChain
	wrapper, err := bc.ReadValidatorInformationAtRoot(addr, block.Root())
	if err != nil {
		return nil, errors.Wrapf(err, "not found address in current state %s", addr)
	}

	now := block.Epoch()
	// At the last block of epoch, block epoch is e while val.LastEpochInCommittee
	// is already updated to e+1. So need the >= check rather than ==
	inCommittee := wrapper.LastEpochInCommittee.Cmp(now) >= 0
	defaultReply := &staking.ValidatorRPCEnhanced{
		CurrentlyInCommittee: inCommittee,
		Wrapper:              *wrapper,
		Performance:          nil,
		ComputedMetrics:      nil,
		TotalDelegated:       wrapper.TotalDelegation(),
		EPoSStatus: effective.ValidatorStatus(
			inCommittee, wrapper.Status,
		).String(),
		EPoSWinningStake: nil,
		BootedStatus:     nil,
		ActiveStatus:     wrapper.Validator.Status.String(),
		Lifetime: &staking.AccumulatedOverLifetime{
			BlockReward: wrapper.BlockReward,
			Signing:     wrapper.Counters,
			APR:         zero,
		},
	}

	snapshot, err := bc.ReadValidatorSnapshotAtEpoch(
		now, addr,
	)

	if err != nil {
		return defaultReply, nil
	}

	computed := availability.ComputeCurrentSigning(
		snapshot.Validator, wrapper,
	)

	lastBlockOfEpoch := shard.Schedule.EpochLastBlock(astra.BeaconChain.CurrentBlock().Header().Epoch().Uint64())

	computed.BlocksLeftInEpoch = lastBlockOfEpoch - astra.BeaconChain.CurrentBlock().Header().Number().Uint64()

	if defaultReply.CurrentlyInCommittee {
		defaultReply.Performance = &staking.CurrentEpochPerformance{
			CurrentSigningPercentage: *computed,
			Epoch:                    astra.BeaconChain.CurrentBlock().Header().Number().Uint64(),
			Block:                    astra.BeaconChain.CurrentBlock().Header().Epoch().Uint64(),
		}
	}

	stats, err := bc.ReadValidatorStats(addr)
	if err != nil {
		// when validator has no stats, default boot-status to not booted
		notBooted := effective.NotBooted.String()
		defaultReply.BootedStatus = &notBooted
		return defaultReply, nil
	}

	latestAPR := numeric.ZeroDec()
	l := len(stats.APRs)
	if l > 0 {
		latestAPR = stats.APRs[l-1].Value
	}
	defaultReply.Lifetime.APR = latestAPR
	defaultReply.Lifetime.EpochAPRs = stats.APRs

	// average apr cache keys
	// key := fmt.Sprintf("apr-%s-%d", addr.Hex(), now.Uint64())
	// prevKey := fmt.Sprintf("apr-%s-%d", addr.Hex(), now.Uint64()-1)

	// delete entry for previous epoch
	// b.apiCache.Forget(prevKey)

	// calculate last APRHistoryLength epochs for averaging APR
	// epochFrom := bc.GasPriceConfig().StakingEpoch
	// nowMinus := big.NewInt(0).Sub(now, big.NewInt(staking.APRHistoryLength))
	// if nowMinus.Cmp(epochFrom) > 0 {
	// 	epochFrom = nowMinus
	// }

	// if len(stats.APRs) > 0 && stats.APRs[0].Epoch.Cmp(epochFrom) > 0 {
	// 	epochFrom = stats.APRs[0].Epoch
	// }

	// epochToAPRs := map[int64]numeric.Dec{}
	// for i := 0; i < len(stats.APRs); i++ {
	// 	entry := stats.APRs[i]
	// 	epochToAPRs[entry.Epoch.Int64()] = entry.Value
	// }

	// at this point, validator is active and has apr's for the recent 100 epochs
	// compute average apr over history
	// if avgAPR, err := b.SingleFlightRequest(
	// 	key, func() (interface{}, error) {
	// 		total := numeric.ZeroDec()
	// 		count := 0
	// 		for i := epochFrom.Int64(); i < now.Int64(); i++ {
	// 			if apr, ok := epochToAPRs[i]; ok {
	// 				total = total.Add(apr)
	// 			}
	// 			count++
	// 		}
	// 		if count == 0 {
	// 			return nil, errors.New("no apr snapshots available")
	// 		}
	// 		return total.QuoInt64(int64(count)), nil
	// 	},
	// ); err != nil {
	// 	// could not compute average apr from snapshot
	// 	// assign the latest apr available from stats
	// 	defaultReply.Lifetime.APR = numeric.ZeroDec()
	// } else {
	// 	defaultReply.Lifetime.APR = avgAPR.(numeric.Dec)
	// }

	epochBlocks := []staking.EpochSigningEntry{}
	epochFrom := bc.Config().StakingEpoch
	nowMinus := big.NewInt(0).Sub(now, big.NewInt(staking.SigningHistoryLength))
	if nowMinus.Cmp(epochFrom) > 0 {
		epochFrom = nowMinus
	}
	for i := now.Int64(); i > epochFrom.Int64(); i-- {
		epoch := big.NewInt(i)
		entry, err := astra.getEpochSigning(epoch, addr)
		if err != nil {
			break
		}
		epochBlocks = append(epochBlocks, entry)
	}
	defaultReply.Lifetime.EpochBlocks = epochBlocks

	if defaultReply.CurrentlyInCommittee {
		defaultReply.ComputedMetrics = stats
		defaultReply.EPoSWinningStake = &stats.TotalEffectiveStake
	}

	if !defaultReply.CurrentlyInCommittee {
		reason := stats.BootedStatus.String()
		defaultReply.BootedStatus = &reason
	}

	return defaultReply, nil
}

// GetMedianRawStakeSnapshot ..
func (astra *Astra) GetMedianRawStakeSnapshot() (
	*committee.CompletedEPoSRound, error,
) {
	blockNum := astra.CurrentBlock().NumberU64()
	key := fmt.Sprintf("median-%d", blockNum)

	// delete cache for previous block
	prevKey := fmt.Sprintf("median-%d", blockNum-1)
	astra.group.Forget(prevKey)

	res, err := astra.SingleFlightRequest(
		key,
		func() (interface{}, error) {
			// Compute for next epoch
			epoch := big.NewInt(0).Add(astra.CurrentBlock().Epoch(), big.NewInt(1))
			instance := shard.Schedule.InstanceForEpoch(epoch)
			return committee.NewEPoSRound(epoch, astra.BlockChain, astra.BlockChain.Config().IsEPoSBound35(epoch), instance.SlotsLimit(), int(instance.NumShards()))
		},
	)
	if err != nil {
		return nil, err
	}
	return res.(*committee.CompletedEPoSRound), nil
}

// GetDelegationsByValidator returns all delegation information of a validator
func (astra *Astra) GetDelegationsByValidator(validator common.Address) []staking.Delegation {
	wrapper, err := astra.BlockChain.ReadValidatorInformation(validator)
	if err != nil || wrapper == nil {
		return nil
	}
	return wrapper.Delegations
}

// GetDelegationsByValidatorAtBlock returns all delegation information of a validator at the given block
func (astra *Astra) GetDelegationsByValidatorAtBlock(
	validator common.Address, block *types.Block,
) []staking.Delegation {
	wrapper, err := astra.BlockChain.ReadValidatorInformationAtRoot(validator, block.Root())
	if err != nil || wrapper == nil {
		return nil
	}
	return wrapper.Delegations
}

// GetDelegationsByDelegator returns all delegation information of a delegator
func (astra *Astra) GetDelegationsByDelegator(
	delegator common.Address,
) ([]common.Address, []*staking.Delegation) {
	block := astra.BlockChain.CurrentBlock()
	return astra.GetDelegationsByDelegatorByBlock(delegator, block)
}

// GetDelegationsByDelegatorByBlock returns all delegation information of a delegator
func (astra *Astra) GetDelegationsByDelegatorByBlock(
	delegator common.Address, block *types.Block,
) ([]common.Address, []*staking.Delegation) {
	delegationIndexes, err := astra.BlockChain.
		ReadDelegationsByDelegatorAt(delegator, block.Number())
	if err != nil {
		return nil, nil
	}

	addresses := make([]common.Address, 0, len(delegationIndexes))
	delegations := make([]*staking.Delegation, 0, len(delegationIndexes))

	for i := range delegationIndexes {
		wrapper, err := astra.BlockChain.ReadValidatorInformationAtRoot(
			delegationIndexes[i].ValidatorAddress, block.Root(),
		)
		if err != nil || wrapper == nil {
			return nil, nil
		}

		if uint64(len(wrapper.Delegations)) > delegationIndexes[i].Index {
			delegations = append(delegations, &wrapper.Delegations[delegationIndexes[i].Index])
		} else {
			delegations = append(delegations, nil)
		}
		addresses = append(addresses, delegationIndexes[i].ValidatorAddress)
	}
	return addresses, delegations
}

// UndelegationPayouts ..
type UndelegationPayouts struct {
	Data map[common.Address]map[common.Address]*big.Int
}

func NewUndelegationPayouts() *UndelegationPayouts {
	return &UndelegationPayouts{
		Data: make(map[common.Address]map[common.Address]*big.Int),
	}
}

func (u *UndelegationPayouts) SetPayoutByDelegatorAddrAndValidatorAddr(
	delegator, validator common.Address, amount *big.Int,
) {
	if u.Data[delegator] == nil {
		u.Data[delegator] = make(map[common.Address]*big.Int)
	}

	if totalPayout, ok := u.Data[delegator][validator]; ok {
		u.Data[delegator][validator] = new(big.Int).Add(totalPayout, amount)
	} else {
		u.Data[delegator][validator] = amount
	}
}

// GetUndelegationPayouts returns the undelegation payouts for each delegator
//
// Due to in-memory caching, it is possible to get undelegation payouts for a state / epoch
// that has been pruned but have it be lost (and unable to recompute) after the node restarts.
// This not a problem if a full (archival) DB is used.
func (astra *Astra) GetUndelegationPayouts(
	ctx context.Context, epoch *big.Int,
) (*UndelegationPayouts, error) {
	if !astra.IsPreStakingEpoch(epoch) {
		return nil, fmt.Errorf("not pre-staking epoch or later")
	}

	payouts, ok := astra.undelegationPayoutsCache.Get(epoch.Uint64())
	if ok {
		return payouts.(*UndelegationPayouts), nil
	}
	undelegationPayouts := NewUndelegationPayouts()
	// require second to last block as saved undelegations are AFTER undelegations are payed out
	blockNumber := shard.Schedule.EpochLastBlock(epoch.Uint64()) - 1
	undelegationPayoutBlock, err := astra.BlockByNumber(ctx, rpc.BlockNumber(blockNumber))
	if err != nil || undelegationPayoutBlock == nil {
		// Block not found, so no undelegationPayouts (not an error)
		return undelegationPayouts, nil
	}

	lockingPeriod := astra.GetDelegationLockingPeriodInEpoch(undelegationPayoutBlock.Epoch())
	for _, validator := range astra.GetAllValidatorAddresses() {
		wrapper, err := astra.BlockChain.ReadValidatorInformationAtRoot(validator, undelegationPayoutBlock.Root())
		if err != nil || wrapper == nil {
			continue // Not a validator at this epoch or unable to fetch validator info because of pruned state.
		}
		noEarlyUnlock := astra.IsNoEarlyUnlockEpoch(epoch)
		for _, delegation := range wrapper.Delegations {
			withdraw := delegation.RemoveUnlockedUndelegations(epoch, wrapper.LastEpochInCommittee, lockingPeriod, noEarlyUnlock)
			if withdraw.Cmp(bigZero) == 1 {
				undelegationPayouts.SetPayoutByDelegatorAddrAndValidatorAddr(validator, delegation.DelegatorAddress, withdraw)
			}
		}
	}

	astra.undelegationPayoutsCache.Add(epoch.Uint64(), undelegationPayouts)
	return undelegationPayouts, nil
}

// GetTotalStakingSnapshot ..
func (astra *Astra) GetTotalStakingSnapshot() *big.Int {
	if stake := astra.totalStakeCache.pop(astra.CurrentBlock().NumberU64()); stake != nil {
		return stake
	}
	currHeight := astra.CurrentBlock().NumberU64()
	candidates := astra.BlockChain.ValidatorCandidates()
	if len(candidates) == 0 {
		stake := big.NewInt(0)
		astra.totalStakeCache.push(currHeight, stake)
		return stake
	}
	stakes := big.NewInt(0)
	for i := range candidates {
		snapshot, _ := astra.BlockChain.ReadValidatorSnapshot(candidates[i])
		validator, _ := astra.BlockChain.ReadValidatorInformation(candidates[i])
		if !committee.IsEligibleForEPoSAuction(
			snapshot, validator,
		) {
			continue
		}
		for i := range validator.Delegations {
			stakes.Add(stakes, validator.Delegations[i].Amount)
		}
	}
	astra.totalStakeCache.push(currHeight, stakes)
	return stakes
}

// GetCurrentStakingErrorSink ..
func (astra *Astra) GetCurrentStakingErrorSink() types.TransactionErrorReports {
	return astra.NodeAPI.ReportStakingErrorSink()
}

// totalStakeCache ..
type totalStakeCache struct {
	sync.Mutex
	cachedBlockHeight uint64
	stake             *big.Int
	// duration is in blocks
	duration uint64
}

// newTotalStakeCache ..
func newTotalStakeCache(duration uint64) *totalStakeCache {
	return &totalStakeCache{
		cachedBlockHeight: 0,
		stake:             nil,
		duration:          duration,
	}
}

func (c *totalStakeCache) push(currBlockHeight uint64, stake *big.Int) {
	c.Lock()
	defer c.Unlock()
	c.cachedBlockHeight = currBlockHeight
	c.stake = stake
}

func (c *totalStakeCache) pop(currBlockHeight uint64) *big.Int {
	if currBlockHeight > c.cachedBlockHeight+c.duration {
		return nil
	}
	return c.stake
}
