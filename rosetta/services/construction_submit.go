package services

import (
	"context"
	"fmt"

	astraTypes "github.com/astra-net/astra-network/core/types"
	"github.com/astra-net/astra-network/rosetta/common"
	stakingTypes "github.com/astra-net/astra-network/staking/types"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/pkg/errors"
)

// ConstructionHash implements the /construction/hash endpoint.
func (s *ConstructAPI) ConstructionHash(
	ctx context.Context, request *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	if err := assertValidNetworkIdentifier(request.NetworkIdentifier, s.astra.ShardID); err != nil {
		return nil, err
	}
	_, tx, rosettaError := unpackWrappedTransactionFromString(request.SignedTransaction, true)
	if rosettaError != nil {
		return nil, rosettaError
	}
	if tx == nil {
		return nil, common.NewError(common.CatchAllError, map[string]interface{}{
			"message": "nil transaction",
		})
	}
	if tx.ShardID() != s.astra.ShardID {
		return nil, common.NewError(common.InvalidTransactionConstructionError, map[string]interface{}{
			"message": fmt.Sprintf("transaction is for shard %v != shard %v", tx.ShardID(), s.astra.ShardID),
		})
	}
	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: tx.Hash().String()},
	}, nil
}

// ConstructionSubmit implements the /construction/submit endpoint.
func (s *ConstructAPI) ConstructionSubmit(
	ctx context.Context, request *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	if err := assertValidNetworkIdentifier(request.NetworkIdentifier, s.astra.ShardID); err != nil {
		return nil, err
	}
	wrappedTransaction, tx, rosettaError := unpackWrappedTransactionFromString(request.SignedTransaction, true)
	if rosettaError != nil {
		return nil, rosettaError
	}
	if wrappedTransaction == nil || tx == nil {
		return nil, common.NewError(common.CatchAllError, map[string]interface{}{
			"message": "nil wrapped transaction or nil unwrapped transaction",
		})
	}
	if tx.ShardID() != s.astra.ShardID {
		return nil, common.NewError(common.StakingTransactionSubmissionError, map[string]interface{}{
			"message": fmt.Sprintf("transaction is for shard %v != shard %v", tx.ShardID(), s.astra.ShardID),
		})
	}

	wrappedSenderAddress, err := getAddress(wrappedTransaction.From)
	if err != nil {
		return nil, common.NewError(common.StakingTransactionSubmissionError, map[string]interface{}{
			"message": errors.WithMessage(err, "unable to get address from wrapped transaction"),
		})
	}

	var signedTx astraTypes.PoolTransaction
	if stakingTx, ok := tx.(*stakingTypes.StakingTransaction); ok && wrappedTransaction.IsStaking {
		signedTx = stakingTx
	} else if plainTx, ok := tx.(*astraTypes.Transaction); ok && !wrappedTransaction.IsStaking {
		signedTx = plainTx
	} else {
		return nil, common.NewError(common.CatchAllError, map[string]interface{}{
			"message": "invalid/inconsistent type or unknown transaction type stored in wrapped transaction",
		})
	}

	txSenderAddress, err := signedTx.SenderAddress()
	if err != nil {
		return nil, common.NewError(common.StakingTransactionSubmissionError, map[string]interface{}{
			"message": errors.WithMessage(err, "unable to get sender address from transaction").Error(),
		})
	}

	if wrappedSenderAddress != txSenderAddress {
		return nil, common.NewError(common.StakingTransactionSubmissionError, map[string]interface{}{
			"message": "transaction sender address does not match wrapped transaction sender address",
		})
	}

	if wrappedTransaction.IsStaking {
		if err := s.astra.SendStakingTx(ctx, signedTx.(*stakingTypes.StakingTransaction)); err != nil {
			return nil, common.NewError(common.StakingTransactionSubmissionError, map[string]interface{}{
				"message": fmt.Sprintf("error is: %s, gas price is: %s, gas limit is: %d", err.Error(), signedTx.GasPrice().String(), signedTx.GasLimit()),
			})
		}
	} else {
		if err := s.astra.SendTx(ctx, signedTx.(*astraTypes.Transaction)); err != nil {
			return nil, common.NewError(common.TransactionSubmissionError, map[string]interface{}{
				"message": err.Error(),
			})
		}
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: tx.Hash().String()},
	}, nil
}
