package downloader

import (
	"context"

	"github.com/Astra-Net/AstraNetwork/consensus/engine"
	"github.com/Astra-Net/AstraNetwork/core/types"
	"github.com/Astra-Net/AstraNetwork/p2p/stream/common/streammanager"
	syncproto "github.com/Astra-Net/AstraNetwork/p2p/stream/protocols/sync"
	sttypes "github.com/Astra-Net/AstraNetwork/p2p/stream/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

type syncProtocol interface {
	GetCurrentBlockNumber(ctx context.Context, opts ...syncproto.Option) (uint64, sttypes.StreamID, error)
	GetBlocksByNumber(ctx context.Context, bns []uint64, opts ...syncproto.Option) ([]*types.Block, sttypes.StreamID, error)
	GetBlockHashes(ctx context.Context, bns []uint64, opts ...syncproto.Option) ([]common.Hash, sttypes.StreamID, error)
	GetBlocksByHashes(ctx context.Context, hs []common.Hash, opts ...syncproto.Option) ([]*types.Block, sttypes.StreamID, error)

	RemoveStream(stID sttypes.StreamID) // If a stream delivers invalid data, remove the stream
	SubscribeAddStreamEvent(ch chan<- streammanager.EvtStreamAdded) event.Subscription
	NumStreams() int
}

type blockChain interface {
	engine.ChainReader
	Engine() engine.Engine

	InsertChain(chain types.Blocks, verifyHeaders bool) (int, error)
	WriteCommitSig(blockNum uint64, lastCommits []byte) error
}
