package astra

import (
	"context"
	"math/big"

	"github.com/astra-net/astra-network/core/types"
	"github.com/ethereum/go-ethereum/common"
)

// GetPoolStats returns the number of pending and queued transactions
func (astra *Astra) GetPoolStats() (pendingCount, queuedCount int) {
	return astra.TxPool.Stats()
}

// GetPoolNonce ...
func (astra *Astra) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return astra.TxPool.State().GetNonce(addr), nil
}

// GetPoolTransaction ...
func (astra *Astra) GetPoolTransaction(hash common.Hash) types.PoolTransaction {
	return astra.TxPool.Get(hash)
}

// GetPendingCXReceipts ..
func (astra *Astra) GetPendingCXReceipts() []*types.CXReceiptsProof {
	return astra.NodeAPI.PendingCXReceipts()
}

// GetPoolTransactions returns pool transactions.
func (astra *Astra) GetPoolTransactions() (types.PoolTransactions, error) {
	pending, err := astra.TxPool.Pending()
	if err != nil {
		return nil, err
	}
	queued, err := astra.TxPool.Queued()
	if err != nil {
		return nil, err
	}
	var txs types.PoolTransactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	for _, batch := range queued {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (astra *Astra) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return astra.gpo.SuggestPrice(ctx)
}
