package node

import (
	"strings"
	"testing"

	"github.com/Astra-Net/AstraNetwork/internal/shardchain"

	"github.com/Astra-Net/AstraNetwork/consensus"
	"github.com/Astra-Net/AstraNetwork/consensus/quorum"
	"github.com/Astra-Net/AstraNetwork/core/types"
	"github.com/Astra-Net/AstraNetwork/crypto/bls"
	"github.com/Astra-Net/AstraNetwork/internal/utils"
	"github.com/Astra-Net/AstraNetwork/multibls"
	"github.com/Astra-Net/AstraNetwork/p2p"
	"github.com/Astra-Net/AstraNetwork/shard"
	staking "github.com/Astra-Net/AstraNetwork/staking/types"
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
		t.Fatalf("Cannot craeate consensus: %v", err)
	}
	var testDBFactory = &shardchain.MemDBFactory{}
	node := New(host, consensus, testDBFactory, nil, nil, nil)

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
