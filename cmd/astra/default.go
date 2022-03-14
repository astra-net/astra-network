package main

import (
	astraconfig "github.com/harmony-one/astra/internal/configs/astra"
	nodeconfig "github.com/harmony-one/astra/internal/configs/node"
)

const tomlConfigVersion = "2.5.1" // bump from 2.5.0 for AccountSlots

const (
	defNetworkType = nodeconfig.Mainnet
)

var defaultConfig = astraconfig.AstraConfig{
	Version: tomlConfigVersion,
	General: astraconfig.GeneralConfig{
		NodeType:         "validator",
		NoStaking:        false,
		ShardID:          -1,
		IsArchival:       false,
		IsBeaconArchival: false,
		IsOffline:        false,
		DataDir:          "./",
	},
	Network: getDefaultNetworkConfig(defNetworkType),
	P2P: astraconfig.P2pConfig{
		Port:            nodeconfig.DefaultP2PPort,
		IP:              nodeconfig.DefaultPublicListenIP,
		KeyFile:         "./.astrakey",
		DiscConcurrency: nodeconfig.DefaultP2PConcurrency,
		MaxConnsPerIP:   nodeconfig.DefaultMaxConnPerIP,
	},
	HTTP: astraconfig.HttpConfig{
		Enabled:        true,
		RosettaEnabled: false,
		IP:             "127.0.0.1",
		Port:           nodeconfig.DefaultRPCPort,
		AuthPort:       nodeconfig.DefaultAuthRPCPort,
		RosettaPort:    nodeconfig.DefaultRosettaPort,
	},
	WS: astraconfig.WsConfig{
		Enabled:  true,
		IP:       "127.0.0.1",
		Port:     nodeconfig.DefaultWSPort,
		AuthPort: nodeconfig.DefaultAuthWSPort,
	},
	RPCOpt: astraconfig.RpcOptConfig{
		DebugEnabled:      false,
		RateLimterEnabled: true,
		RequestsPerSecond: nodeconfig.DefaultRPCRateLimit,
	},
	BLSKeys: astraconfig.BlsConfig{
		KeyDir:   "./.astra/blskeys",
		KeyFiles: []string{},
		MaxKeys:  10,

		PassEnabled:      true,
		PassSrcType:      blsPassTypeAuto,
		PassFile:         "",
		SavePassphrase:   false,
		KMSEnabled:       false,
		KMSConfigSrcType: kmsConfigTypeShared,
		KMSConfigFile:    "",
	},
	TxPool: astraconfig.TxPoolConfig{
		BlacklistFile:  "./.astra/blacklist.txt",
		RosettaFixFile: "",
		AccountSlots:   16,
	},
	Sync: getDefaultSyncConfig(defNetworkType),
	Pprof: astraconfig.PprofConfig{
		Enabled:            false,
		ListenAddr:         "127.0.0.1:6060",
		Folder:             "./profiles",
		ProfileNames:       []string{},
		ProfileIntervals:   []int{600},
		ProfileDebugValues: []int{0},
	},
	Log: astraconfig.LogConfig{
		Folder:       "./latest",
		FileName:     "astra.log",
		RotateSize:   100,
		RotateCount:  0,
		RotateMaxAge: 0,
		Verbosity:    3,
		VerbosePrints: astraconfig.LogVerbosePrints{
			Config: true,
		},
	},
	DNSSync: getDefaultDNSSyncConfig(defNetworkType),
	ShardData: astraconfig.ShardDataConfig{
		EnableShardData: false,
		DiskCount:       8,
		ShardCount:      4,
		CacheTime:       10,
		CacheSize:       512,
	},
}

var defaultSysConfig = astraconfig.SysConfig{
	NtpServer: "1.pool.ntp.org",
}

var defaultDevnetConfig = astraconfig.DevnetConfig{
	NumShards:     2,
	ShardSize:     10,
	AstraNodeSize: 10,
}

var defaultRevertConfig = astraconfig.RevertConfig{
	RevertBeacon: false,
	RevertBefore: 0,
	RevertTo:     0,
}

var defaultLogContext = astraconfig.LogContext{
	IP:   "127.0.0.1",
	Port: 9000,
}

var defaultConsensusConfig = astraconfig.ConsensusConfig{
	MinPeers:     6,
	AggregateSig: true,
}

var defaultPrometheusConfig = astraconfig.PrometheusConfig{
	Enabled:    true,
	IP:         "0.0.0.0",
	Port:       9900,
	EnablePush: false,
	Gateway:    "https://gateway.astranetwork.com",
}

var (
	defaultMainnetSyncConfig = astraconfig.SyncConfig{
		Enabled:        false,
		Downloader:     false,
		Concurrency:    6,
		MinPeers:       6,
		InitStreams:    8,
		DiscSoftLowCap: 8,
		DiscHardLowCap: 6,
		DiscHighCap:    128,
		DiscBatch:      8,
	}

	defaultTestNetSyncConfig = astraconfig.SyncConfig{
		Enabled:        true,
		Downloader:     false,
		Concurrency:    2,
		MinPeers:       2,
		InitStreams:    2,
		DiscSoftLowCap: 2,
		DiscHardLowCap: 2,
		DiscHighCap:    1024,
		DiscBatch:      3,
	}

	defaultLocalNetSyncConfig = astraconfig.SyncConfig{
		Enabled:        true,
		Downloader:     true,
		Concurrency:    2,
		MinPeers:       2,
		InitStreams:    2,
		DiscSoftLowCap: 2,
		DiscHardLowCap: 2,
		DiscHighCap:    1024,
		DiscBatch:      3,
	}

	defaultElseSyncConfig = astraconfig.SyncConfig{
		Enabled:        true,
		Downloader:     true,
		Concurrency:    4,
		MinPeers:       4,
		InitStreams:    4,
		DiscSoftLowCap: 4,
		DiscHardLowCap: 4,
		DiscHighCap:    1024,
		DiscBatch:      8,
	}
)

const (
	defaultBroadcastInvalidTx = false
)

func getDefaultAstraConfigCopy(nt nodeconfig.NetworkType) astraconfig.AstraConfig {
	config := defaultConfig

	config.Network = getDefaultNetworkConfig(nt)
	if nt == nodeconfig.Devnet {
		devnet := getDefaultDevnetConfigCopy()
		config.Devnet = &devnet
	}
	config.Sync = getDefaultSyncConfig(nt)
	config.DNSSync = getDefaultDNSSyncConfig(nt)

	return config
}

func getDefaultSysConfigCopy() astraconfig.SysConfig {
	config := defaultSysConfig
	return config
}

func getDefaultDevnetConfigCopy() astraconfig.DevnetConfig {
	config := defaultDevnetConfig
	return config
}

func getDefaultRevertConfigCopy() astraconfig.RevertConfig {
	config := defaultRevertConfig
	return config
}

func getDefaultLogContextCopy() astraconfig.LogContext {
	config := defaultLogContext
	return config
}

func getDefaultConsensusConfigCopy() astraconfig.ConsensusConfig {
	config := defaultConsensusConfig
	return config
}

func getDefaultPrometheusConfigCopy() astraconfig.PrometheusConfig {
	config := defaultPrometheusConfig
	return config
}

const (
	nodeTypeValidator = "validator"
	nodeTypeExplorer  = "explorer"
)

const (
	blsPassTypeAuto   = "auto"
	blsPassTypeFile   = "file"
	blsPassTypePrompt = "prompt"

	kmsConfigTypeShared = "shared"
	kmsConfigTypePrompt = "prompt"
	kmsConfigTypeFile   = "file"

	legacyBLSPassTypeDefault = "default"
	legacyBLSPassTypeStdin   = "stdin"
	legacyBLSPassTypeDynamic = "no-prompt"
	legacyBLSPassTypePrompt  = "prompt"
	legacyBLSPassTypeStatic  = "file"
	legacyBLSPassTypeNone    = "none"

	legacyBLSKmsTypeDefault = "default"
	legacyBLSKmsTypePrompt  = "prompt"
	legacyBLSKmsTypeFile    = "file"
	legacyBLSKmsTypeNone    = "none"
)
