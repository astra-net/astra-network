package slash

import (
	"math/big"

	"github.com/harmony-one/astra/core/types"
	"github.com/harmony-one/astra/internal/params"
	"github.com/harmony-one/astra/shard"
)

// CommitteeReader ..
type CommitteeReader interface {
	Config() *params.ChainConfig
	ReadShardState(epoch *big.Int) (*shard.State, error)
	CurrentBlock() *types.Block
}
