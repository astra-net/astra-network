package v2

import (
	"context"
	"math/big"

	"github.com/astra-net/astra-network/eth/rpc"
	"github.com/astra-net/astra-network/astra"
	internal_common "github.com/astra-net/astra-network/internal/common"
)

// PublicLegacyService provides an API to access the Astra blockchain.
// Services here are legacy methods, specific to the V1 RPC that can be deprecated in the future.
type PublicLegacyService struct {
	astra *astra.Astra
}

// NewPublicLegacyAPI creates a new API for the RPC interface
func NewPublicLegacyAPI(astra *astra.Astra, namespace string) rpc.API {
	if namespace == "" {
		namespace = "astrav2"
	}

	return rpc.API{
		Namespace: namespace,
		Version:   "1.0",
		Service:   &PublicLegacyService{astra},
		Public:    true,
	}
}

// GetBalance returns the amount of Atto for the given address in the state of the
// given block number.
func (s *PublicLegacyService) GetBalance(
	ctx context.Context, address string,
) (*big.Int, error) {
	addr, err := internal_common.ParseAddr(address)
	if err != nil {
		return nil, err
	}
	balance, err := s.astra.GetBalance(ctx, addr, rpc.BlockNumber(-1))
	if err != nil {
		return nil, err
	}
	return balance, nil
}
