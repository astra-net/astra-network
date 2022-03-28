package rpc

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/astra-net/astra-network/astra"
	"github.com/astra-net/astra-network/eth/rpc"
	nodeconfig "github.com/astra-net/astra-network/internal/configs/node"
	"github.com/astra-net/astra-network/internal/utils"
	eth "github.com/astra-net/astra-network/rpc/eth"
	v1 "github.com/astra-net/astra-network/rpc/v1"
	v2 "github.com/astra-net/astra-network/rpc/v2"
)

// Version enum
const (
	V1 Version = iota
	V2
	Eth
	Debug
	Trace
)

const (
	// APIVersion used for DApp's, bumped after RPC refactor (7/2020)
	APIVersion = "1.1"
	// CallTimeout is the timeout given to all contract calls
	CallTimeout = 5 * time.Second
	// LogTag is the tag found in the log for all RPC logs
	LogTag = "[RPC]"
	// HTTPPortOffset ..
	HTTPPortOffset = 500
	// WSPortOffset ..
	WSPortOffset = 800

	netNamespace   = "net"
	netV1Namespace = "netv1"
	netV2Namespace = "netv2"
	web3Namespace  = "web3"
)

var (
	// HTTPModules ..
	HTTPModules = []string{"astra", "astrav2", "eth", "debug", "trace", netNamespace, netV1Namespace, netV2Namespace, web3Namespace, "explorer"}
	// WSModules ..
	WSModules = []string{"astra", "astrav2", "eth", "debug", "trace", netNamespace, netV1Namespace, netV2Namespace, web3Namespace, "web3"}

	httpListener     net.Listener
	httpHandler      *rpc.Server
	wsListener       net.Listener
	wsHandler        *rpc.Server
	httpEndpoint     = ""
	httpAuthEndpoint = ""
	wsEndpoint       = ""
	wsAuthEndpoint   = ""
	httpVirtualHosts = []string{"*"}
	httpTimeouts     = rpc.DefaultHTTPTimeouts
	httpOrigins      = []string{"*"}
	wsOrigins        = []string{"*"}
)

// Version of the RPC
type Version int

// Namespace of the RPC version
func (n Version) Namespace() string {
	return HTTPModules[n]
}

// StartServers starts the http & ws servers
func StartServers(astra *astra.Astra, apis []rpc.API, config nodeconfig.RPCServerConfig) error {
	apis = append(apis, getAPIs(astra, config.DebugEnabled, config.RateLimiterEnabled, config.RequestsPerSecond)...)
	authApis := append(apis, getAuthAPIs(astra, config.DebugEnabled, config.RateLimiterEnabled, config.RequestsPerSecond)...)

	if config.HTTPEnabled {
		httpEndpoint = fmt.Sprintf("%v:%v", config.HTTPIp, config.HTTPPort)
		if err := startHTTP(apis); err != nil {
			return err
		}

		httpAuthEndpoint = fmt.Sprintf("%v:%v", config.HTTPIp, config.HTTPAuthPort)
		if err := startAuthHTTP(authApis); err != nil {
			return err
		}
	}

	if config.WSEnabled {
		wsEndpoint = fmt.Sprintf("%v:%v", config.WSIp, config.WSPort)
		if err := startWS(apis); err != nil {
			return err
		}

		wsAuthEndpoint = fmt.Sprintf("%v:%v", config.WSIp, config.WSAuthPort)
		if err := startAuthWS(authApis); err != nil {
			return err
		}
	}

	return nil
}

// StopServers stops the http & ws servers
func StopServers() error {
	if httpListener != nil {
		if err := httpListener.Close(); err != nil {
			return err
		}
		httpListener = nil
		utils.Logger().Info().
			Str("url", fmt.Sprintf("http://%s", httpEndpoint)).
			Msg("HTTP endpoint closed")
	}
	if httpHandler != nil {
		httpHandler.Stop()
		httpHandler = nil
	}
	if wsListener != nil {
		if err := wsListener.Close(); err != nil {
			return err
		}
		wsListener = nil
		utils.Logger().Info().
			Str("url", fmt.Sprintf("http://%s", wsEndpoint)).
			Msg("WS endpoint closed")
	}
	if wsHandler != nil {
		wsHandler.Stop()
		wsHandler = nil
	}
	return nil
}

func getAuthAPIs(astra *astra.Astra, debugEnable bool, rateLimiterEnable bool, ratelimit int) []rpc.API {
	return []rpc.API{
		NewPublicTraceAPI(astra, Debug), // Debug version means geth trace rpc
		NewPublicTraceAPI(astra, Trace), // Trace version means parity trace rpc
	}
}

// getAPIs returns all the API methods for the RPC interface
func getAPIs(astra *astra.Astra, debugEnable bool, rateLimiterEnable bool, ratelimit int) []rpc.API {
	publicAPIs := []rpc.API{
		// Public methods
		NewPublicAstraAPI(astra, V1),
		NewPublicAstraAPI(astra, V2),
		NewPublicAstraAPI(astra, Eth),
		NewPublicBlockchainAPI(astra, V1, rateLimiterEnable, ratelimit),
		NewPublicBlockchainAPI(astra, V2, rateLimiterEnable, ratelimit),
		NewPublicBlockchainAPI(astra, Eth, rateLimiterEnable, ratelimit),
		NewPublicContractAPI(astra, V1),
		NewPublicContractAPI(astra, V2),
		NewPublicContractAPI(astra, Eth),
		NewPublicTransactionAPI(astra, V1),
		NewPublicTransactionAPI(astra, V2),
		NewPublicTransactionAPI(astra, Eth),
		NewPublicPoolAPI(astra, V1),
		NewPublicPoolAPI(astra, V2),
		NewPublicPoolAPI(astra, Eth),
		NewPublicStakingAPI(astra, V1),
		NewPublicStakingAPI(astra, V2),
		NewPublicDebugAPI(astra, V1),
		NewPublicDebugAPI(astra, V2),
		// Legacy methods (subject to removal)
		v1.NewPublicLegacyAPI(astra, "astra"),
		eth.NewPublicEthService(astra, "eth"),
		v2.NewPublicLegacyAPI(astra, "astrav2"),
	}

	publicDebugAPIs := []rpc.API{
		//Public debug API
		NewPublicDebugAPI(astra, V1),
		NewPublicDebugAPI(astra, V2),
	}

	privateAPIs := []rpc.API{
		NewPrivateDebugAPI(astra, V1),
		NewPrivateDebugAPI(astra, V2),
	}

	if debugEnable {
		apis := append(publicAPIs, publicDebugAPIs...)
		return append(apis, privateAPIs...)
	}
	return publicAPIs
}

func startHTTP(apis []rpc.API) (err error) {
	httpListener, httpHandler, err = rpc.StartHTTPEndpoint(
		httpEndpoint, apis, HTTPModules, httpOrigins, httpVirtualHosts, httpTimeouts,
	)
	if err != nil {
		return err
	}

	utils.Logger().Info().
		Str("url", fmt.Sprintf("http://%s", httpEndpoint)).
		Str("cors", strings.Join(httpOrigins, ",")).
		Str("vhosts", strings.Join(httpVirtualHosts, ",")).
		Msg("HTTP endpoint opened")
	fmt.Printf("Started RPC server at: %v\n", httpEndpoint)
	return nil
}

func startAuthHTTP(apis []rpc.API) (err error) {
	httpListener, httpHandler, err = rpc.StartHTTPEndpoint(
		httpAuthEndpoint, apis, HTTPModules, httpOrigins, httpVirtualHosts, httpTimeouts,
	)
	if err != nil {
		return err
	}

	utils.Logger().Info().
		Str("url", fmt.Sprintf("http://%s", httpAuthEndpoint)).
		Str("cors", strings.Join(httpOrigins, ",")).
		Str("vhosts", strings.Join(httpVirtualHosts, ",")).
		Msg("HTTP endpoint opened")
	fmt.Printf("Started Auth-RPC server at: %v\n", httpAuthEndpoint)
	return nil
}

func startWS(apis []rpc.API) (err error) {
	wsListener, wsHandler, err = rpc.StartWSEndpoint(wsEndpoint, apis, WSModules, wsOrigins, true)
	if err != nil {
		return err
	}

	utils.Logger().Info().
		Str("url", fmt.Sprintf("ws://%s", wsListener.Addr())).
		Msg("WebSocket endpoint opened")
	fmt.Printf("Started WS server at: %v\n", wsEndpoint)
	return nil
}

func startAuthWS(apis []rpc.API) (err error) {
	wsListener, wsHandler, err = rpc.StartWSEndpoint(wsAuthEndpoint, apis, WSModules, wsOrigins, true)
	if err != nil {
		return err
	}

	utils.Logger().Info().
		Str("url", fmt.Sprintf("ws://%s", wsListener.Addr())).
		Msg("WebSocket endpoint opened")
	fmt.Printf("Started Auth-WS server at: %v\n", wsAuthEndpoint)
	return nil
}
