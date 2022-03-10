package hmy

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/harmony-one/astra/core/types"
)

// GetPoolStats returns the number of pending and queued transactions
func (hmy *Astra) GetPoolStats() (pendingCount, queuedCount int) {
	return hmy.TxPool.Stats()
}

// GetPoolNonce ...
func (hmy *Astra) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return hmy.TxPool.State().GetNonce(addr), nil
}

// GetPoolTransaction ...
func (hmy *Astra) GetPoolTransaction(hash common.Hash) types.PoolTransaction {
	return hmy.TxPool.Get(hash)
}

// GetPendingCXReceipts ..
func (hmy *Astra) GetPendingCXReceipts() []*types.CXReceiptsProof {
	return hmy.NodeAPI.PendingCXReceipts()
}

// GetPoolTransactions returns pool transactions.
func (hmy *Astra) GetPoolTransactions() (types.PoolTransactions, error) {
	pending, err := hmy.TxPool.Pending()
	if err != nil {
		return nil, err
	}
	queued, err := hmy.TxPool.Queued()
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

func (hmy *Astra) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return hmy.gpo.SuggestPrice(ctx)
}
