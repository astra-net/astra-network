package core

import (
	"math/big"
	"testing"

	blockfactory "github.com/astra-net/astra-network/block/factory"
	"github.com/astra-net/astra-network/core/types"
	shardingconfig "github.com/astra-net/astra-network/internal/configs/sharding"
	"github.com/astra-net/astra-network/shard"
)

func TestIsEpochBlock(t *testing.T) {
	blockNumbered := func(n int64) *types.Block {
		return types.NewBlock(
			blockfactory.NewTestHeader().With().Number(big.NewInt(n)).Header(),
			nil, nil, nil, nil, nil,
		)
	}
	tests := []struct {
		schedule shardingconfig.Schedule
		block    *types.Block
		expected bool
	}{
		{
			shardingconfig.MainnetSchedule,
			blockNumbered(10),
			false,
		},
		{
			shardingconfig.MainnetSchedule,
			blockNumbered(0),
			true,
		},
		{
			shardingconfig.MainnetSchedule,
			blockNumbered(344064),
			true,
		},
		{
			shardingconfig.TestnetSchedule,
			blockNumbered(37),
			false,
		},
		{
			shardingconfig.TestnetSchedule,
			blockNumbered(38),
			false,
		},
		{
			shardingconfig.TestnetSchedule,
			blockNumbered(75),
			false,
		},
		{
			shardingconfig.TestnetSchedule,
			blockNumbered(76),
			false,
		},
	}
	for i, test := range tests {
		shard.Schedule = test.schedule
		r := IsEpochBlock(test.block)
		if r != test.expected {
			t.Errorf("index: %v, expected: %v, got: %v\n", i, test.expected, r)
		}
	}
}
