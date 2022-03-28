package slash

import (
	"math/big"

	"github.com/astra-net/AstraNetwork/core/types"
	"github.com/astra-net/AstraNetwork/internal/params"
	"github.com/astra-net/AstraNetwork/shard"
)

// CommitteeReader ..
type CommitteeReader interface {
	Config() *params.ChainConfig
	ReadShardState(epoch *big.Int) (*shard.State, error)
	CurrentBlock() *types.Block
}
