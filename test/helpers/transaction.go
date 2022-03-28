package helpers

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/coinbase/rosetta-sdk-go/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	astratypes "github.com/Astra-Net/AstraNetwork/core/types"
	rpcV2 "github.com/Astra-Net/AstraNetwork/rpc/v2"
	stakingTypes "github.com/Astra-Net/AstraNetwork/staking/types"
)

// CreateTestStakingTransaction creates a pre-signed staking transaction
func CreateTestStakingTransaction(
	payloadMaker func() (stakingTypes.Directive, interface{}), key *ecdsa.PrivateKey,
	nonce, gasLimit uint64, gasPrice *big.Int,
) (*stakingTypes.StakingTransaction, error) {
	tx, err := stakingTypes.NewStakingTransaction(nonce, gasLimit, gasPrice, payloadMaker)
	if err != nil {
		return nil, err
	}
	if key == nil {
		key, err = crypto.GenerateKey()
		if err != nil {
			return nil, err
		}
	}
	// Staking transactions are always post EIP155 epoch
	return stakingTypes.Sign(tx, stakingTypes.NewEIP155Signer(tx.ChainID()), key)
}

// GetMessageFromStakingTx gets the staking message, as seen by the rpc layer
func GetMessageFromStakingTx(tx *stakingTypes.StakingTransaction) (map[string]interface{}, error) {
	rpcStakingTx, err := rpcV2.NewStakingTransaction(tx, ethcommon.Hash{}, 0, 0, 0, true)
	if err != nil {
		return nil, err
	}
	return types.MarshalMap(rpcStakingTx.Msg)
}

// CreateTestTransaction creates a pre-signed transaction
func CreateTestTransaction(
	signer astratypes.Signer, fromShard, toShard uint32, nonce, gasLimit uint64,
	gasPrice, amount *big.Int, data []byte,
) (*astratypes.Transaction, error) {
	fromKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	toKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	toAddr := crypto.PubkeyToAddress(toKey.PublicKey)
	var tx *astratypes.Transaction
	if fromShard != toShard {
		tx = astratypes.NewCrossShardTransaction(
			nonce, &toAddr, fromShard, toShard, amount, gasLimit, gasPrice, data,
		)
	} else {
		tx = astratypes.NewTransaction(
			nonce, toAddr, fromShard, amount, gasLimit, gasPrice, data,
		)
	}
	return astratypes.SignTx(tx, signer, fromKey)
}

// CreateTestContractCreationTransaction creates a pre-signed contract creation transaction
func CreateTestContractCreationTransaction(
	signer astratypes.Signer, shard uint32, nonce, gasLimit uint64, gasPrice, amount *big.Int, data []byte,
) (*astratypes.Transaction, error) {
	fromKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	tx := astratypes.NewContractCreation(nonce, shard, amount, gasLimit, gasPrice, data)
	return astratypes.SignTx(tx, signer, fromKey)
}
