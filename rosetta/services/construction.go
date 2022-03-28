package services

import (
	"context"
	"fmt"
	"math/big"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/astra-net/AstraNetwork/astra"
	"github.com/astra-net/AstraNetwork/common/denominations"
	astraTypes "github.com/astra-net/AstraNetwork/core/types"
	"github.com/astra-net/AstraNetwork/rosetta/common"
	stakingTypes "github.com/astra-net/AstraNetwork/staking/types"
)

const (
	// DefaultGasPrice ..
	DefaultGasPrice = 30 * denominations.Nano
)

// ConstructAPI implements the server.ConstructAPIServicer interface.
type ConstructAPI struct {
	astra           *astra.Astra
	signer        astraTypes.Signer
	stakingSigner stakingTypes.Signer
}

// NewConstructionAPI creates a new instance of a ConstructAPI.
func NewConstructionAPI(astra *astra.Astra) server.ConstructionAPIServicer {
	return &ConstructAPI{
		astra:           astra,
		signer:        astraTypes.NewEIP155Signer(new(big.Int).SetUint64(astra.ChainID)),
		stakingSigner: stakingTypes.NewEIP155Signer(new(big.Int).SetUint64(astra.ChainID)),
	}
}

// ConstructionDerive implements the /construction/derive endpoint.
func (s *ConstructAPI) ConstructionDerive(
	ctx context.Context, request *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	if err := assertValidNetworkIdentifier(request.NetworkIdentifier, s.astra.ShardID); err != nil {
		return nil, err
	}
	address, rosettaError := getAddressFromPublicKey(request.PublicKey)
	if rosettaError != nil {
		return nil, rosettaError
	}
	accountID, rosettaError := newAccountIdentifier(*address)
	if rosettaError != nil {
		return nil, rosettaError
	}
	return &types.ConstructionDeriveResponse{
		AccountIdentifier: accountID,
	}, nil
}

// getAddressFromPublicKey assumes that data is a compressed secp256k1 public key
func getAddressFromPublicKey(
	key *types.PublicKey,
) (*ethCommon.Address, *types.Error) {
	if key == nil {
		return nil, common.NewError(common.CatchAllError, map[string]interface{}{
			"message": "nil key",
		})
	}
	if key.CurveType != common.CurveType {
		return nil, common.NewError(common.UnsupportedCurveTypeError, map[string]interface{}{
			"message": fmt.Sprintf("currently only support %v", common.CurveType),
		})
	}
	// underlying eth crypto lib uses secp256k1
	publicKey, err := crypto.DecompressPubkey(key.Bytes)
	if err != nil {
		return nil, common.NewError(common.CatchAllError, map[string]interface{}{
			"message": err.Error(),
		})
	}
	address := crypto.PubkeyToAddress(*publicKey)
	return &address, nil
}
