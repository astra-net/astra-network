package slash

import (
	"math/big"

	"github.com/Astra-Net/AstraNetwork/core/types"
	"github.com/Astra-Net/AstraNetwork/internal/params"
	"github.com/Astra-Net/AstraNetwork/shard"
)

// CommitteeReader ..
type CommitteeReader interface {
	Config() *params.ChainConfig
	ReadShardState(epoch *big.Int) (*shard.State, error)
	CurrentBlock() *types.Block
}
