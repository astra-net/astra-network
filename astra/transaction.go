package astra

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/harmony-one/astra/core"
	"github.com/harmony-one/astra/core/rawdb"
	"github.com/harmony-one/astra/core/types"
	"github.com/harmony-one/astra/eth/rpc"
)

// SendTx ...
func (astra *Astra) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	tx, _, _, _ := rawdb.ReadTransaction(astra.chainDb, signedTx.Hash())
	if tx == nil {
		return astra.NodeAPI.AddPendingTransaction(signedTx)
	}
	return ErrFinalizedTransaction
}

// ResendCx retrieve blockHash from txID and add blockHash to CxPool for resending
// Note that cross shard txn is only for regular txns, not for staking txns, so the input txn hash
// is expected to be regular txn hash
func (astra *Astra) ResendCx(ctx context.Context, txID common.Hash) (uint64, bool) {
	blockHash, blockNum, index := astra.BlockChain.ReadTxLookupEntry(txID)
	if blockHash == (common.Hash{}) {
		return 0, false
	}

	blk := astra.BlockChain.GetBlockByHash(blockHash)
	if blk == nil {
		return 0, false
	}

	txs := blk.Transactions()
	// a valid index is from 0 to len-1
	if int(index) > len(txs)-1 {
		return 0, false
	}
	tx := txs[int(index)]

	// check whether it is a valid cross shard tx
	if tx.ShardID() == tx.ToShardID() || blk.Header().ShardID() != tx.ShardID() {
		return 0, false
	}
	entry := core.CxEntry{blockHash, tx.ToShardID()}
	success := astra.CxPool.Add(entry)
	return blockNum, success
}

// GetReceipts ...
func (astra *Astra) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return astra.BlockChain.GetReceiptsByHash(hash), nil
}

// GetTransactionsHistory returns list of transactions hashes of address.
func (astra *Astra) GetTransactionsHistory(address, txType, order string) ([]common.Hash, error) {
	return astra.NodeAPI.GetTransactionsHistory(address, txType, order)
}

// GetAccountNonce returns the nonce value of the given address for the given block number
func (astra *Astra) GetAccountNonce(
	ctx context.Context, address common.Address, blockNum rpc.BlockNumber) (uint64, error) {
	state, _, err := astra.StateAndHeaderByNumber(ctx, blockNum)
	if state == nil || err != nil {
		return 0, err
	}
	return state.GetNonce(address), state.Error()
}

// GetTransactionsCount returns the number of regular transactions of address.
func (astra *Astra) GetTransactionsCount(address, txType string) (uint64, error) {
	return astra.NodeAPI.GetTransactionsCount(address, txType)
}

// GetCurrentTransactionErrorSink ..
func (astra *Astra) GetCurrentTransactionErrorSink() types.TransactionErrorReports {
	return astra.NodeAPI.ReportPlainErrorSink()
}
