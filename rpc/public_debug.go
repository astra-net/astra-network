package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
	"github.com/harmony-one/astra/eth/rpc"
	"github.com/harmony-one/astra/astra"
	"github.com/harmony-one/astra/internal/utils"
)

// PublicDebugService Internal JSON RPC for debugging purpose
type PublicDebugService struct {
	astra     *astra.Astra
	version Version
}

// NewPublicDebugAPI creates a new API for the RPC interface
func NewPublicDebugAPI(astra *astra.Astra, version Version) rpc.API {
	return rpc.API{
		Namespace: version.Namespace(),
		Version:   APIVersion,
		Service:   &PublicDebugService{astra, version},
		Public:    false,
	}
}

// SetLogVerbosity Sets log verbosity on runtime
// curl -H "Content-Type: application/json" -d '{"method":"astra_setLogVerbosity","params":[5],"id":1}' http://127.0.0.1:9500
func (s *PublicDebugService) SetLogVerbosity(ctx context.Context, level int) (map[string]interface{}, error) {
	if level < int(log.LvlCrit) || level > int(log.LvlTrace) {
		return nil, ErrInvalidLogLevel
	}

	verbosity := log.Lvl(level)
	utils.SetLogVerbosity(verbosity)
	return map[string]interface{}{"verbosity": verbosity.String()}, nil
}
