package common

import (
	"github.com/astra-net/astra-network/consensus/quorum"
	"github.com/astra-net/astra-network/crypto/bls"
	"github.com/astra-net/astra-network/numeric"
)

type setRawStakeHack interface {
	SetRawStake(key bls.SerializedPublicKey, d numeric.Dec)
}

// SetRawStake is a hack, return value is if was successful or not at setting
func SetRawStake(q quorum.Decider, key bls.SerializedPublicKey, d numeric.Dec) bool {
	if setter, ok := q.(setRawStakeHack); ok {
		setter.SetRawStake(key, d)
		return true
	}
	return false
}
