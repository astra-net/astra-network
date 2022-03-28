package common

import (
	"github.com/astra-net/AstraNetwork/consensus/quorum"
	"github.com/astra-net/AstraNetwork/crypto/bls"
	"github.com/astra-net/AstraNetwork/numeric"
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
