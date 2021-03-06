package eth

import (
	"context"

	"github.com/astra-net/astra-network/astra"
	"github.com/astra-net/astra-network/eth/rpc"
	internal_common "github.com/astra-net/astra-network/internal/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// PublicEthService provides an API to access to the Eth endpoints for the Astra blockchain.
type PublicEthService struct {
	astra *astra.Astra
}

// NewPublicEthService creates a new API for the RPC interface
func NewPublicEthService(astra *astra.Astra, namespace string) rpc.API {
	if namespace == "" {
		namespace = "eth"
	}

	return rpc.API{
		Namespace: namespace,
		Version:   "1.0",
		Service:   &PublicEthService{astra},
		Public:    true,
	}
}

// GetBalance returns the amount of Atto for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicEthService) GetBalance(
	ctx context.Context, address string, blockNrOrHash rpc.BlockNumberOrHash,
) (*hexutil.Big, error) {
	addr, err := internal_common.ParseAddr(address)
	if err != nil {
		return nil, err
	}
	balance, err := s.astra.GetBalance(ctx, addr, blockNrOrHash)
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(balance), nil
}
