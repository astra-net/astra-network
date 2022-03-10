package v1

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/harmony-one/astra/eth/rpc"
	"github.com/harmony-one/astra/astra"
	internal_common "github.com/harmony-one/astra/internal/common"
)

// PublicLegacyService provides an API to access the Astra blockchain.
// Services here are legacy methods, specific to the V1 RPC that can be deprecated in the future.
type PublicLegacyService struct {
	astra *astra.Astra
}

// NewPublicLegacyAPI creates a new API for the RPC interface
func NewPublicLegacyAPI(astra *astra.Astra, namespace string) rpc.API {
	if namespace == "" {
		namespace = "astra"
	}

	return rpc.API{
		Namespace: namespace,
		Version:   "1.0",
		Service:   &PublicLegacyService{astra},
		Public:    true,
	}
}

// GetBalance returns the amount of Atto for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicLegacyService) GetBalance(
	ctx context.Context, address string, blockNr rpc.BlockNumber,
) (*hexutil.Big, error) {
	addr, err := internal_common.ParseAddr(address)
	if err != nil {
		return nil, err
	}
	balance, err := s.astra.GetBalance(ctx, addr, blockNr)
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(balance), nil
}
