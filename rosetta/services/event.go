package services

import (
	"context"

	"github.com/Astra-Net/AstraNetwork/astra"
	astraTypes "github.com/Astra-Net/AstraNetwork/core/types"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// EventAPI implements the server.EventsAPIServicer interface.
type EventAPI struct {
	astra *astra.Astra
}

func NewEventAPI(astra *astra.Astra) *EventAPI {
	return &EventAPI{astra: astra}
}

// EventsBlocks implements the /events/blocks endpoint
func (e *EventAPI) EventsBlocks(ctx context.Context, request *types.EventsBlocksRequest) (resp *types.EventsBlocksResponse, err *types.Error) {
	cacheItem, cacheHelper, cacheErr := rosettaCacheHelper("EventsBlocks", request)
	if cacheErr == nil {
		if cacheItem != nil {
			return cacheItem.resp.(*types.EventsBlocksResponse), nil
		} else {
			defer cacheHelper(resp, err)
		}
	}

	if err := assertValidNetworkIdentifier(request.NetworkIdentifier, e.astra.ShardID); err != nil {
		return nil, err
	}

	var offset, limit int64

	if request.Limit == nil {
		limit = 10
	} else {
		limit = *request.Limit
		if limit > 1000 {
			limit = 1000
		}
	}

	if request.Offset == nil {
		offset = 0
	} else {
		offset = *request.Offset
	}

	resp = &types.EventsBlocksResponse{
		MaxSequence: e.astra.BlockChain.CurrentHeader().Number().Int64(),
	}

	for i := offset; i < offset+limit; i++ {
		block := e.astra.BlockChain.GetBlockByNumber(uint64(i))
		if block == nil {
			break
		}

		resp.Events = append(resp.Events, buildFromBlock(block))
	}

	return resp, nil
}

func buildFromBlock(block *astraTypes.Block) *types.BlockEvent {
	return &types.BlockEvent{
		Sequence: block.Number().Int64(),
		BlockIdentifier: &types.BlockIdentifier{
			Index: block.Number().Int64(),
			Hash:  block.Hash().Hex(),
		},
		Type: types.ADDED,
	}
}
