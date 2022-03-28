package slash

import (
	"math/big"

	"github.com/astra-net/astra-network/core/types"
	"github.com/astra-net/astra-network/internal/params"
	"github.com/astra-net/astra-network/shard"
)

// CommitteeReader ..
type CommitteeReader interface {
	Config() *params.ChainConfig
	ReadShardState(epoch *big.Int) (*shard.State, error)
	CurrentBlock() *types.Block
}
