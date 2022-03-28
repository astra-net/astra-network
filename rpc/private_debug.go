package rpc

import (
	"context"

	"github.com/Astra-Net/AstraNetwork/astra"
	"github.com/Astra-Net/AstraNetwork/eth/rpc"
)

// PrivateDebugService Internal JSON RPC for debugging purpose
type PrivateDebugService struct {
	astra     *astra.Astra
	version Version
}

// NewPrivateDebugAPI creates a new API for the RPC interface
// TODO(dm): expose public via config
func NewPrivateDebugAPI(astra *astra.Astra, version Version) rpc.API {
	return rpc.API{
		Namespace: version.Namespace(),
		Version:   APIVersion,
		Service:   &PrivateDebugService{astra, version},
		Public:    false,
	}
}

// ConsensusViewChangingID return the current view changing ID to RPC
func (s *PrivateDebugService) ConsensusViewChangingID(
	ctx context.Context,
) uint64 {
	return s.astra.NodeAPI.GetConsensusViewChangingID()
}

// ConsensusCurViewID return the current view ID to RPC
func (s *PrivateDebugService) ConsensusCurViewID(
	ctx context.Context,
) uint64 {
	return s.astra.NodeAPI.GetConsensusCurViewID()
}

// GetConsensusMode return the current consensus mode
func (s *PrivateDebugService) GetConsensusMode(
	ctx context.Context,
) string {
	return s.astra.NodeAPI.GetConsensusMode()
}

// GetConsensusPhase return the current consensus mode
func (s *PrivateDebugService) GetConsensusPhase(
	ctx context.Context,
) string {
	return s.astra.NodeAPI.GetConsensusPhase()
}

// GetConfig get astra config
func (s *PrivateDebugService) GetConfig(
	ctx context.Context,
) (StructuredResponse, error) {
	return NewStructuredResponse(s.astra.NodeAPI.GetConfig())
}

// GetLastSigningPower get last signed power
func (s *PrivateDebugService) GetLastSigningPower(
	ctx context.Context,
) (float64, error) {
	return s.astra.NodeAPI.GetLastSigningPower()
}
