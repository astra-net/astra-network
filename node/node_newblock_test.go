package node

import (
	"strings"
	"testing"

	"github.com/astra-net/astra-network/internal/shardchain"

	"github.com/astra-net/astra-network/consensus"
	"github.com/astra-net/astra-network/consensus/quorum"
	"github.com/astra-net/astra-network/core/types"
	"github.com/astra-net/astra-network/crypto/bls"
	"github.com/astra-net/astra-network/internal/utils"
	"github.com/astra-net/astra-network/multibls"
	"github.com/astra-net/astra-network/p2p"
	"github.com/astra-net/astra-network/shard"
	staking "github.com/astra-net/astra-network/staking/types"
	"github.com/ethereum/go-ethereum/common"
)

func TestFinalizeNewBlockAsync(t *testing.T) {
	blsKey := bls.RandPrivateKey()
	pubKey := blsKey.GetPublicKey()
	leader := p2p.Peer{IP: "127.0.0.1", Port: "8882", ConsensusPubKey: pubKey}
	priKey, _, _ := utils.GenKeyP2P("127.0.0.1", "9902")
	host, err := p2p.NewHost(p2p.HostConfig{
		Self:   &leader,
		BLSKey: priKey,
	})
	if err != nil {
		t.Fatalf("newhost failure: %v", err)
	}
	decider := quorum.NewDecider(
		quorum.SuperMajorityVote, shard.BeaconChainShardID,
	)
	consensus, err := consensus.New(
		host, shard.BeaconChainShardID, leader, multibls.GetPrivateKeys(blsKey), decider,
	)
	if err != nil {
		t.Fatalf("Cannot create consensus: %v", err)
	}
	var testDBFactory = &shardchain.MemDBFactory{}
	node := New(host, consensus, testDBFactory, nil, nil, nil, nil)

	node.Worker.UpdateCurrent()

	txs := make(map[common.Address]types.Transactions)
	stks := staking.StakingTransactions{}
	node.Worker.CommitTransactions(
		txs, stks, common.Address{},
	)
	commitSigs := make(chan []byte)
	go func() {
		commitSigs <- []byte{}
	}()

	block, _ := node.Worker.FinalizeNewBlock(
		commitSigs, func() uint64 { return 0 }, common.Address{}, nil, nil,
	)

	if err := node.VerifyNewBlock(block); err != nil {
		t.Error("New block is not verified successfully:", err)
	}

	node.Blockchain().InsertChain(types.Blocks{block}, false)

	node.Worker.UpdateCurrent()

	_, err = node.Worker.FinalizeNewBlock(
		commitSigs, func() uint64 { return 0 }, common.Address{}, nil, nil,
	)

	if !strings.Contains(err.Error(), "cannot finalize block") {
		t.Error("expect timeout on FinalizeNewBlock")
	}
}
