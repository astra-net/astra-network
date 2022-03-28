package rpc

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/astra-net/astra-network/astra"
	"github.com/astra-net/astra-network/eth/rpc"
	"github.com/ethereum/go-ethereum/common"
)

var (
	parityTraceGO = "ParityBlockTracer"
)

type PublicParityTracerService struct {
	*PublicTracerService
}

func (s *PublicParityTracerService) Transaction(ctx context.Context, hash common.Hash) (interface{}, error) {
	timer := DoMetricRPCRequest(Transaction)
	defer DoRPCRequestDuration(Transaction, timer)
	return s.TraceTransaction(ctx, hash, &astra.TraceConfig{Tracer: &parityTraceGO})
}

// trace_block RPC
func (s *PublicParityTracerService) Block(ctx context.Context, number rpc.BlockNumber) (interface{}, error) {
	timer := DoMetricRPCRequest(Block)
	defer DoRPCRequestDuration(Block, timer)

	block := s.astra.BlockChain.GetBlockByNumber(uint64(number))
	if block == nil {
		return nil, nil
	}
	results, err := s.astra.TraceBlock(ctx, block, &astra.TraceConfig{Tracer: &parityTraceGO})
	if err != nil {
		return results, err
	}
	var resultArray = make([]json.RawMessage, 0)
	for _, result := range results {
		raw, ok := result.Result.([]json.RawMessage)
		if !ok {
			return results, errors.New("tracer bug:expected []json.RawMessage")
		}
		resultArray = append(resultArray, raw...)
	}
	return resultArray, nil
}
