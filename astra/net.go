package astra

import (
	nodeconfig "github.com/astra-net/AstraNetwork/internal/configs/node"
	commonRPC "github.com/astra-net/AstraNetwork/rpc/common"
	"github.com/astra-net/AstraNetwork/staking/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

// GetCurrentUtilityMetrics ..
func (astra *Astra) GetCurrentUtilityMetrics() (*network.UtilityMetric, error) {
	return network.NewUtilityMetricSnapshot(astra.BlockChain)
}

// GetPeerInfo returns the peer info to the node, including blocked peer, connected peer, number of peers
func (astra *Astra) GetPeerInfo() commonRPC.NodePeerInfo {

	topics := astra.NodeAPI.ListTopic()
	p := make([]commonRPC.P, len(topics))

	for i, t := range topics {
		topicPeer := astra.NodeAPI.ListPeer(t)
		p[i].Topic = t
		p[i].Peers = make([]peer.ID, len(topicPeer))
		copy(p[i].Peers, topicPeer)
	}

	return commonRPC.NodePeerInfo{
		PeerID:       nodeconfig.GetPeerID(),
		BlockedPeers: astra.NodeAPI.ListBlockedPeer(),
		P:            p,
	}
}
