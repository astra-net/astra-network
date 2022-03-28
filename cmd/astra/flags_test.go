package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	astraconfig "github.com/astra-net/astra-network/internal/configs/astra"

	"github.com/spf13/cobra"

	"github.com/astra-net/astra-network/internal/cli"
	nodeconfig "github.com/astra-net/astra-network/internal/configs/node"
)

var (
	trueBool = true
)

func TestAstraFlags(t *testing.T) {
	tests := []struct {
		argStr    string
		expConfig astraconfig.AstraConfig
	}{
		{
			// running staking command from legacy node.sh
			argStr: "--bootnodes /ip4/100.26.90.187/tcp/9874/p2p/Qmdfjtk6hPoyrH1zVD9PEH4zfWLo38dP2mDvvKXfh3tnEv," +
				"/ip4/54.213.43.194/tcp/9874/p2p/QmZJJx6AdaoEkGLrYG4JeLCKeCKDjnFz2wfHNHxAqFSGA9,/ip4/13.113.101." +
				"219/tcp/12019/p2p/QmQayinFSgMMw5cSpDUiD9pQ2WeP6WNmGxpZ6ou3mdVFJX,/ip4/99.81.170.167/tcp/12019/p" +
				"2p/QmRVbTpEYup8dSaURZfF6ByrMTSKa4UyUzJhSjahFzRqNj --ip 8.8.8.8 --port 9000 --network_type=mainn" +
				"et --dns_zone=t.astranetwork.com --blacklist=./.astra/blacklist.txt --min_peers=6 --max_bls_keys_per_node=" +
				"10 --broadcast_invalid_tx=true --verbosity=3 --is_archival=false --shard_id=-1 --staking=true -" +
				"-aws-config-source file:config.json --p2p.disc.concurrency 5 --p2p.security.max-conn-per-ip 5",
			expConfig: astraconfig.AstraConfig{
				Version: tomlConfigVersion,
				General: astraconfig.GeneralConfig{
					NodeType:   "validator",
					NoStaking:  false,
					ShardID:    -1,
					IsArchival: false,
					DataDir:    "./",
				},
				Network: astraconfig.NetworkConfig{
					NetworkType: "mainnet",
					BootNodes: []string{
						"/ip4/100.26.90.187/tcp/9874/p2p/Qmdfjtk6hPoyrH1zVD9PEH4zfWLo38dP2mDvvKXfh3tnEv",
						"/ip4/54.213.43.194/tcp/9874/p2p/QmZJJx6AdaoEkGLrYG4JeLCKeCKDjnFz2wfHNHxAqFSGA9",
						"/ip4/13.113.101.219/tcp/12019/p2p/QmQayinFSgMMw5cSpDUiD9pQ2WeP6WNmGxpZ6ou3mdVFJX",
						"/ip4/99.81.170.167/tcp/12019/p2p/QmRVbTpEYup8dSaURZfF6ByrMTSKa4UyUzJhSjahFzRqNj",
					},
				},
				DNSSync: astraconfig.DnsSync{
					Port:       6000,
					Zone:       "t.astranetwork.com",
					Server:     true,
					Client:     true,
					ServerPort: nodeconfig.DefaultDNSPort,
				},
				P2P: astraconfig.P2pConfig{
					Port:            9000,
					IP:              defaultConfig.P2P.IP,
					KeyFile:         defaultConfig.P2P.KeyFile,
					DiscConcurrency: 5,
					MaxConnsPerIP:   5,
				},
				HTTP: astraconfig.HttpConfig{
					Enabled:        true,
					IP:             "127.0.0.1",
					Port:           9500,
					AuthPort:       9501,
					RosettaEnabled: false,
					RosettaPort:    9700,
				},
				RPCOpt: astraconfig.RpcOptConfig{
					DebugEnabled:      false,
					RateLimterEnabled: true,
					RequestsPerSecond: 1000,
				},
				WS: astraconfig.WsConfig{
					Enabled:  true,
					IP:       "127.0.0.1",
					Port:     9800,
					AuthPort: 9801,
				},
				Consensus: &astraconfig.ConsensusConfig{
					MinPeers:     6,
					AggregateSig: true,
				},
				BLSKeys: astraconfig.BlsConfig{
					KeyDir:           "./.astra/blskeys",
					KeyFiles:         []string{},
					MaxKeys:          10,
					PassEnabled:      true,
					PassSrcType:      "auto",
					PassFile:         "",
					SavePassphrase:   false,
					KMSEnabled:       false,
					KMSConfigSrcType: "file",
					KMSConfigFile:    "config.json",
				},
				TxPool: astraconfig.TxPoolConfig{
					BlacklistFile:  "./.astra/blacklist.txt",
					RosettaFixFile: "",
					AccountSlots:   16,
				},
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
					FileName:     "validator-8.8.8.8-9000.log",
					RotateSize:   100,
					RotateCount:  0,
					RotateMaxAge: 0,
					Verbosity:    3,
					Context: &astraconfig.LogContext{
						IP:   "8.8.8.8",
						Port: 9000,
					},
					VerbosePrints: astraconfig.LogVerbosePrints{
						Config: true,
					},
				},
				Sys: &astraconfig.SysConfig{
					NtpServer: defaultSysConfig.NtpServer,
				},
				Legacy: &astraconfig.LegacyConfig{
					TPBroadcastInvalidTxn: &trueBool,
				},
				Prometheus: &astraconfig.PrometheusConfig{
					Enabled:    true,
					IP:         "0.0.0.0",
					Port:       9900,
					EnablePush: true,
					Gateway:    "https://gateway.astranetwork.com",
				},
				Sync: defaultMainnetSyncConfig,
				ShardData: astraconfig.ShardDataConfig{
					EnableShardData: false,
					DiskCount:       8,
					ShardCount:      4,
					CacheTime:       10,
					CacheSize:       512,
				},
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, getRootFlags(), applyRootFlags)
		hc, err := ts.run(strings.Split(test.argStr, " "))
		if err != nil {
			t.Fatalf("Test %v: %v", i, err)
		}
		if !reflect.DeepEqual(hc, test.expConfig) {
			t.Errorf("Test %v: unexpected config: \n\t%+v\n\t%+v", i, hc, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestGeneralFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.GeneralConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: astraconfig.GeneralConfig{
				NodeType:   "validator",
				NoStaking:  false,
				ShardID:    -1,
				IsArchival: false,
				DataDir:    "./",
			},
		},
		{
			args: []string{"--run", "explorer", "--run.legacy", "--run.shard=0",
				"--run.archive=true", "--datadir=./.astra"},
			expConfig: astraconfig.GeneralConfig{
				NodeType:   "explorer",
				NoStaking:  true,
				ShardID:    0,
				IsArchival: true,
				DataDir:    "./.astra",
			},
		},
		{
			args: []string{"--node_type", "explorer", "--staking", "--shard_id", "0",
				"--is_archival", "--db_dir", "./"},
			expConfig: astraconfig.GeneralConfig{
				NodeType:   "explorer",
				NoStaking:  false,
				ShardID:    0,
				IsArchival: true,
				DataDir:    "./",
			},
		},
		{
			args: []string{"--staking=false", "--is_archival=false"},
			expConfig: astraconfig.GeneralConfig{
				NodeType:   "validator",
				NoStaking:  true,
				ShardID:    -1,
				IsArchival: false,
				DataDir:    "./",
			},
		},
		{
			args: []string{"--run", "explorer", "--run.shard", "0"},
			expConfig: astraconfig.GeneralConfig{
				NodeType:   "explorer",
				NoStaking:  false,
				ShardID:    0,
				IsArchival: false,
				DataDir:    "./",
			},
		},
		{
			args: []string{"--run", "explorer", "--run.shard", "0", "--run.archive=false"},
			expConfig: astraconfig.GeneralConfig{
				NodeType:   "explorer",
				NoStaking:  false,
				ShardID:    0,
				IsArchival: false,
				DataDir:    "./",
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, generalFlags, applyGeneralFlags)

		got, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(got.General, test.expConfig) {
			t.Errorf("Test %v: unexpected config: \n\t%+v\n\t%+v", i, got.General, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestNetworkFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.AstraConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: astraconfig.AstraConfig{
				Network: astraconfig.NetworkConfig{
					NetworkType: defNetworkType,
					BootNodes:   nodeconfig.GetDefaultBootNodes(defNetworkType),
				},
				DNSSync: getDefaultDNSSyncConfig(defNetworkType)},
		},
		{
			args: []string{"-n", "stn"},
			expConfig: astraconfig.AstraConfig{
				Network: astraconfig.NetworkConfig{
					NetworkType: nodeconfig.Stressnet,
					BootNodes:   nodeconfig.GetDefaultBootNodes(nodeconfig.Stressnet),
				},
				DNSSync: getDefaultDNSSyncConfig(nodeconfig.Stressnet),
			},
		},
		{
			args: []string{"--network", "stk", "--bootnodes", "1,2,3,4", "--dns.zone", "8.8.8.8",
				"--dns.port", "9001", "--dns.server-port", "9002"},
			expConfig: astraconfig.AstraConfig{
				Network: astraconfig.NetworkConfig{
					NetworkType: "pangaea",
					BootNodes:   []string{"1", "2", "3", "4"},
				},
				DNSSync: astraconfig.DnsSync{
					Port:          9001,
					Zone:          "8.8.8.8",
					LegacySyncing: false,
					Server:        true,
					ServerPort:    9002,
				},
			},
		},
		{
			args: []string{"--network_type", "stk", "--bootnodes", "1,2,3,4", "--dns_zone", "8.8.8.8",
				"--dns_port", "9001"},
			expConfig: astraconfig.AstraConfig{
				Network: astraconfig.NetworkConfig{
					NetworkType: "pangaea",
					BootNodes:   []string{"1", "2", "3", "4"},
				},
				DNSSync: astraconfig.DnsSync{
					Port:          9001,
					Zone:          "8.8.8.8",
					LegacySyncing: false,
					Server:        true,
					ServerPort:    nodeconfig.GetDefaultDNSPort(nodeconfig.Pangaea),
				},
			},
		},
		{
			args: []string{"--dns=false"},
			expConfig: astraconfig.AstraConfig{
				Network: astraconfig.NetworkConfig{
					NetworkType: defNetworkType,
					BootNodes:   nodeconfig.GetDefaultBootNodes(defNetworkType),
				},
				DNSSync: astraconfig.DnsSync{
					Port:          nodeconfig.GetDefaultDNSPort(defNetworkType),
					Zone:          nodeconfig.GetDefaultDNSZone(defNetworkType),
					LegacySyncing: true,
					Client:        true,
					Server:        true,
					ServerPort:    nodeconfig.GetDefaultDNSPort(nodeconfig.Pangaea),
				},
			},
		},
	}
	for i, test := range tests {
		neededFlags := make([]cli.Flag, 0)
		neededFlags = append(neededFlags, networkFlags...)
		neededFlags = append(neededFlags, dnsSyncFlags...)
		ts := newFlagTestSuite(t, neededFlags, func(cmd *cobra.Command, config *astraconfig.AstraConfig) {
			// This is the network related logic in function getastraconfig.AstraConfig
			nt := getNetworkType(cmd)
			config.Network = getDefaultNetworkConfig(nt)
			config.DNSSync = getDefaultDNSSyncConfig(nt)
			applyNetworkFlags(cmd, config)
			applyDNSSyncFlags(cmd, config)
		})

		got, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(got.Network, test.expConfig.Network) {
			t.Errorf("Test %v: unexpected network config: \n\t%+v\n\t%+v", i, got.Network, test.expConfig.Network)
		}
		if !reflect.DeepEqual(got.DNSSync, test.expConfig.DNSSync) {
			t.Errorf("Test %v: unexpected dnssync config: \n\t%+v\n\t%+v", i, got.DNSSync, test.expConfig.DNSSync)
		}
		ts.tearDown()
	}
}

var defDataStore = ".dht-127.0.0.1"

func TestP2PFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.P2pConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.P2P,
		},
		{
			args: []string{"--p2p.port", "9001", "--p2p.keyfile", "./key.file", "--p2p.dht.datastore",
				defDataStore},
			expConfig: astraconfig.P2pConfig{
				Port:          9001,
				IP:            nodeconfig.DefaultPublicListenIP,
				KeyFile:       "./key.file",
				DHTDataStore:  &defDataStore,
				MaxConnsPerIP: 10,
			},
		},
		{
			args: []string{"--port", "9001", "--key", "./key.file"},
			expConfig: astraconfig.P2pConfig{
				Port:          9001,
				IP:            nodeconfig.DefaultPublicListenIP,
				KeyFile:       "./key.file",
				MaxConnsPerIP: 10,
			},
		},
		{
			args: []string{"--p2p.port", "9001", "--p2p.disc.concurrency", "5", "--p2p.security.max-conn-per-ip", "5"},
			expConfig: astraconfig.P2pConfig{
				Port:            9001,
				IP:              nodeconfig.DefaultPublicListenIP,
				KeyFile:         "./.astrakey",
				DiscConcurrency: 5,
				MaxConnsPerIP:   5,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, append(p2pFlags, legacyMiscFlags...),
			func(cmd *cobra.Command, config *astraconfig.AstraConfig) {
				applyLegacyMiscFlags(cmd, config)
				applyP2PFlags(cmd, config)
			},
		)

		got, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(got.P2P, test.expConfig) {
			t.Errorf("Test %v: unexpected config: \n\t%+v\n\t%+v", i, got.P2P, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestRPCFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.HttpConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.HTTP,
		},
		{
			args: []string{"--http=false"},
			expConfig: astraconfig.HttpConfig{
				Enabled:        false,
				RosettaEnabled: false,
				IP:             defaultConfig.HTTP.IP,
				Port:           defaultConfig.HTTP.Port,
				AuthPort:       defaultConfig.HTTP.AuthPort,
				RosettaPort:    defaultConfig.HTTP.RosettaPort,
			},
		},
		{
			args: []string{"--http.ip", "8.8.8.8", "--http.port", "9001"},
			expConfig: astraconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: false,
				IP:             "8.8.8.8",
				Port:           9001,
				AuthPort:       defaultConfig.HTTP.AuthPort,
				RosettaPort:    defaultConfig.HTTP.RosettaPort,
			},
		},
		{
			args: []string{"--http.ip", "8.8.8.8", "--http.auth-port", "9001"},
			expConfig: astraconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: false,
				IP:             "8.8.8.8",
				Port:           defaultConfig.HTTP.Port,
				AuthPort:       9001,
				RosettaPort:    defaultConfig.HTTP.RosettaPort,
			},
		},
		{
			args: []string{"--http.ip", "8.8.8.8", "--http.port", "9001", "--http.rosetta.port", "10001"},
			expConfig: astraconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: true,
				IP:             "8.8.8.8",
				Port:           9001,
				AuthPort:       defaultConfig.HTTP.AuthPort,
				RosettaPort:    10001,
			},
		},
		{
			args: []string{"--http.ip", "8.8.8.8", "--http.rosetta.port", "10001"},
			expConfig: astraconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: true,
				IP:             "8.8.8.8",
				Port:           defaultConfig.HTTP.Port,
				AuthPort:       defaultConfig.HTTP.AuthPort,
				RosettaPort:    10001,
			},
		},
		{
			args: []string{"--ip", "8.8.8.8", "--port", "9001", "--public_rpc"},
			expConfig: astraconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: false,
				IP:             nodeconfig.DefaultPublicListenIP,
				Port:           9501,
				AuthPort:       9502,
				RosettaPort:    9701,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, append(httpFlags, legacyMiscFlags...),
			func(cmd *cobra.Command, config *astraconfig.AstraConfig) {
				applyLegacyMiscFlags(cmd, config)
				applyHTTPFlags(cmd, config)
			},
		)

		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.HTTP, test.expConfig) {
			t.Errorf("Test %v: unexpected config: \n\t%+v\n\t%+v", i, hc.HTTP, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestWSFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.WsConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.WS,
		},
		{
			args: []string{"--ws=false"},
			expConfig: astraconfig.WsConfig{
				Enabled:  false,
				IP:       defaultConfig.WS.IP,
				Port:     defaultConfig.WS.Port,
				AuthPort: defaultConfig.WS.AuthPort,
			},
		},
		{
			args: []string{"--ws", "--ws.ip", "8.8.8.8", "--ws.port", "9001"},
			expConfig: astraconfig.WsConfig{
				Enabled:  true,
				IP:       "8.8.8.8",
				Port:     9001,
				AuthPort: defaultConfig.WS.AuthPort,
			},
		},
		{
			args: []string{"--ws", "--ws.ip", "8.8.8.8", "--ws.auth-port", "9001"},
			expConfig: astraconfig.WsConfig{
				Enabled:  true,
				IP:       "8.8.8.8",
				Port:     defaultConfig.WS.Port,
				AuthPort: 9001,
			},
		},
		{
			args: []string{"--ip", "8.8.8.8", "--port", "9001", "--public_rpc"},
			expConfig: astraconfig.WsConfig{
				Enabled:  true,
				IP:       nodeconfig.DefaultPublicListenIP,
				Port:     9801,
				AuthPort: 9802,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, append(wsFlags, legacyMiscFlags...),
			func(cmd *cobra.Command, config *astraconfig.AstraConfig) {
				applyLegacyMiscFlags(cmd, config)
				applyWSFlags(cmd, config)
			},
		)

		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.WS, test.expConfig) {
			t.Errorf("Test %v: \n\t%+v\n\t%+v", i, hc.WS, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestRPCOptFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.RpcOptConfig
	}{
		{
			args: []string{"--rpc.debug"},
			expConfig: astraconfig.RpcOptConfig{
				DebugEnabled:      true,
				RateLimterEnabled: true,
				RequestsPerSecond: 1000,
			},
		},

		{
			args: []string{},
			expConfig: astraconfig.RpcOptConfig{
				DebugEnabled:      false,
				RateLimterEnabled: true,
				RequestsPerSecond: 1000,
			},
		},

		{
			args: []string{"--rpc.ratelimiter", "--rpc.ratelimit", "2000"},
			expConfig: astraconfig.RpcOptConfig{
				DebugEnabled:      false,
				RateLimterEnabled: true,
				RequestsPerSecond: 2000,
			},
		},

		{
			args: []string{"--rpc.ratelimiter=false", "--rpc.ratelimit", "2000"},
			expConfig: astraconfig.RpcOptConfig{
				DebugEnabled:      false,
				RateLimterEnabled: false,
				RequestsPerSecond: 2000,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, rpcOptFlags, applyRPCOptFlags)

		hc, _ := ts.run(test.args)

		if !reflect.DeepEqual(hc.RPCOpt, test.expConfig) {
			t.Errorf("Test %v: \n\t%+v\n\t%+v", i, hc.RPCOpt, test.expConfig)
		}

		ts.tearDown()
	}
}

func TestBLSFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.BlsConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.BLSKeys,
		},
		{
			args: []string{"--bls.dir", "./blskeys", "--bls.keys", "key1,key2",
				"--bls.maxkeys", "8", "--bls.pass", "--bls.pass.src", "auto", "--bls.pass.save",
				"--bls.kms", "--bls.kms.src", "shared",
			},
			expConfig: astraconfig.BlsConfig{
				KeyDir:           "./blskeys",
				KeyFiles:         []string{"key1", "key2"},
				MaxKeys:          8,
				PassEnabled:      true,
				PassSrcType:      "auto",
				PassFile:         "",
				SavePassphrase:   true,
				KMSEnabled:       true,
				KMSConfigSrcType: "shared",
				KMSConfigFile:    "",
			},
		},
		{
			args: []string{"--bls.pass.file", "xxx.pass", "--bls.kms.config", "config.json"},
			expConfig: astraconfig.BlsConfig{
				KeyDir:           defaultConfig.BLSKeys.KeyDir,
				KeyFiles:         defaultConfig.BLSKeys.KeyFiles,
				MaxKeys:          defaultConfig.BLSKeys.MaxKeys,
				PassEnabled:      true,
				PassSrcType:      "file",
				PassFile:         "xxx.pass",
				SavePassphrase:   false,
				KMSEnabled:       false,
				KMSConfigSrcType: "file",
				KMSConfigFile:    "config.json",
			},
		},
		{
			args: []string{"--blskey_file", "key1,key2", "--blsfolder", "./astrakeys",
				"--max_bls_keys_per_node", "5", "--blspass", "file:xxx.pass", "--save-passphrase",
				"--aws-config-source", "file:config.json",
			},
			expConfig: astraconfig.BlsConfig{
				KeyDir:           "./astrakeys",
				KeyFiles:         []string{"key1", "key2"},
				MaxKeys:          5,
				PassEnabled:      true,
				PassSrcType:      "file",
				PassFile:         "xxx.pass",
				SavePassphrase:   true,
				KMSEnabled:       false,
				KMSConfigSrcType: "file",
				KMSConfigFile:    "config.json",
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, blsFlags, applyBLSFlags)

		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.BLSKeys, test.expConfig) {
			t.Errorf("Test %v: \n\t%+v\n\t%+v", i, hc.BLSKeys, test.expConfig)
		}

		ts.tearDown()
	}
}

func TestConsensusFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig *astraconfig.ConsensusConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: nil,
		},
		{
			args: []string{"--consensus.min-peers", "10", "--consensus.aggregate-sig=false"},
			expConfig: &astraconfig.ConsensusConfig{
				MinPeers:     10,
				AggregateSig: false,
			},
		},
		{
			args: []string{"--delay_commit", "10ms", "--block_period", "5", "--min_peers", "10",
				"--consensus.aggregate-sig=true"},
			expConfig: &astraconfig.ConsensusConfig{
				MinPeers:     10,
				AggregateSig: true,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, consensusFlags, applyConsensusFlags)

		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.Consensus, test.expConfig) {
			t.Errorf("Test %v: unexpected config \n\t%+v\n\t%+v", i, hc.Consensus, test.expConfig)
		}

		ts.tearDown()
	}
}

func TestTxPoolFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.TxPoolConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: astraconfig.TxPoolConfig{
				BlacklistFile:  defaultConfig.TxPool.BlacklistFile,
				RosettaFixFile: defaultConfig.TxPool.RosettaFixFile,
				AccountSlots:   defaultConfig.TxPool.AccountSlots,
			},
		},
		{
			args: []string{"--txpool.blacklist", "blacklist.file", "--txpool.rosettafixfile", "rosettafix.file"},
			expConfig: astraconfig.TxPoolConfig{
				BlacklistFile:  "blacklist.file",
				RosettaFixFile: "rosettafix.file",
				AccountSlots:   16, // default
			},
		},
		{
			args: []string{"--blacklist", "blacklist.file", "--txpool.rosettafixfile", "rosettafix.file"},
			expConfig: astraconfig.TxPoolConfig{
				BlacklistFile:  "blacklist.file",
				RosettaFixFile: "rosettafix.file",
				AccountSlots:   16, // default
			},
		},
		{
			args: []string{"--txpool.accountslots", "5", "--txpool.blacklist", "blacklist.file", "--txpool.rosettafixfile", "rosettafix.file"},
			expConfig: astraconfig.TxPoolConfig{
				AccountSlots:   5,
				BlacklistFile:  "blacklist.file",
				RosettaFixFile: "rosettafix.file",
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, txPoolFlags, applyTxPoolFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.TxPool, test.expConfig) {
			t.Errorf("Test %v: unexpected config\n\t%+v\n\t%+v", i, hc.TxPool, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestPprofFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.PprofConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.Pprof,
		},
		{
			args: []string{"--pprof"},
			expConfig: astraconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         defaultConfig.Pprof.ListenAddr,
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof.addr", "8.8.8.8:9001"},
			expConfig: astraconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         "8.8.8.8:9001",
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof=false", "--pprof.addr", "8.8.8.8:9001"},
			expConfig: astraconfig.PprofConfig{
				Enabled:            false,
				ListenAddr:         "8.8.8.8:9001",
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof.profile.names", "cpu,heap,mutex"},
			expConfig: astraconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         defaultConfig.Pprof.ListenAddr,
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       []string{"cpu", "heap", "mutex"},
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof.profile.intervals", "0,1"},
			expConfig: astraconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         defaultConfig.Pprof.ListenAddr,
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   []int{0, 1},
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof.profile.debug", "0,1,0"},
			expConfig: astraconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         defaultConfig.Pprof.ListenAddr,
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: []int{0, 1, 0},
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, pprofFlags, applyPprofFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.Pprof, test.expConfig) {
			t.Errorf("Test %v: unexpected config\n\t%+v\n\t%+v", i, hc.Pprof, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestLogFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.LogConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.Log,
		},
		{
			args: []string{"--log.dir", "latest_log", "--log.max-size", "10", "--log.rotate-count", "3",
				"--log.rotate-max-age", "0", "--log.name", "astra.log", "--log.verb", "5",
				"--log.verbose-prints", "config"},
			expConfig: astraconfig.LogConfig{
				Folder:       "latest_log",
				FileName:     "astra.log",
				RotateSize:   10,
				RotateCount:  3,
				RotateMaxAge: 0,
				Verbosity:    5,
				VerbosePrints: astraconfig.LogVerbosePrints{
					Config: true,
				},
				Context: nil,
			},
		},
		{
			args: []string{"--log.ctx.ip", "8.8.8.8", "--log.ctx.port", "9001"},
			expConfig: astraconfig.LogConfig{
				Folder:        defaultConfig.Log.Folder,
				FileName:      defaultConfig.Log.FileName,
				RotateSize:    defaultConfig.Log.RotateSize,
				RotateCount:   defaultConfig.Log.RotateCount,
				RotateMaxAge:  defaultConfig.Log.RotateMaxAge,
				Verbosity:     defaultConfig.Log.Verbosity,
				VerbosePrints: defaultConfig.Log.VerbosePrints,
				Context: &astraconfig.LogContext{
					IP:   "8.8.8.8",
					Port: 9001,
				},
			},
		},
		{
			args: []string{"--log_folder", "latest_log", "--log_max_size", "10", "--verbosity",
				"5", "--ip", "8.8.8.8", "--port", "9001"},
			expConfig: astraconfig.LogConfig{
				Folder:        "latest_log",
				FileName:      "validator-8.8.8.8-9001.log",
				RotateSize:    10,
				RotateCount:   0,
				RotateMaxAge:  0,
				Verbosity:     5,
				VerbosePrints: defaultConfig.Log.VerbosePrints,
				Context: &astraconfig.LogContext{
					IP:   "8.8.8.8",
					Port: 9001,
				},
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, append(logFlags, legacyMiscFlags...),
			func(cmd *cobra.Command, config *astraconfig.AstraConfig) {
				applyLegacyMiscFlags(cmd, config)
				applyLogFlags(cmd, config)
			},
		)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.Log, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Log, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestSysFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig *astraconfig.SysConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: &astraconfig.SysConfig{
				NtpServer: defaultSysConfig.NtpServer,
			},
		},
		{
			args: []string{"--sys.ntp", "0.pool.ntp.org"},
			expConfig: &astraconfig.SysConfig{
				NtpServer: "0.pool.ntp.org",
			},
		},
	}

	for i, test := range tests {
		ts := newFlagTestSuite(t, sysFlags, applySysFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.Sys, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Sys, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestDevnetFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig *astraconfig.DevnetConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: nil,
		},
		{
			args: []string{"--devnet.num-shard", "3", "--devnet.shard-size", "100",
				"--devnet.astra-node-size", "60"},
			expConfig: &astraconfig.DevnetConfig{
				NumShards:     3,
				ShardSize:     100,
				AstraNodeSize: 60,
			},
		},
		{
			args: []string{"--dn_num_shards", "3", "--dn_shard_size", "100", "--dn_astra_size",
				"60"},
			expConfig: &astraconfig.DevnetConfig{
				NumShards:     3,
				ShardSize:     100,
				AstraNodeSize: 60,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, devnetFlags, applyDevnetFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.Devnet, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Devnet, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestRevertFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig *astraconfig.RevertConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: nil,
		},
		{
			args: []string{"--revert.beacon"},
			expConfig: &astraconfig.RevertConfig{
				RevertBeacon: true,
				RevertTo:     defaultRevertConfig.RevertTo,
				RevertBefore: defaultRevertConfig.RevertBefore,
			},
		},
		{
			args: []string{"--revert.beacon", "--revert.to", "100", "--revert.do-before", "10000"},
			expConfig: &astraconfig.RevertConfig{
				RevertBeacon: true,
				RevertTo:     100,
				RevertBefore: 10000,
			},
		},
		{
			args: []string{"--revert_beacon", "--do_revert_before", "10000", "--revert_to", "100"},
			expConfig: &astraconfig.RevertConfig{
				RevertBeacon: true,
				RevertTo:     100,
				RevertBefore: 10000,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, revertFlags, applyRevertFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.Revert, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Revert, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestDNSSyncFlags(t *testing.T) {
	tests := []struct {
		args      []string
		network   string
		expConfig astraconfig.DnsSync
		expErr    error
	}{
		{
			args:      []string{},
			network:   "mainnet",
			expConfig: getDefaultDNSSyncConfig(nodeconfig.Mainnet),
		},
		{
			args:      []string{"--sync.legacy.server", "--sync.legacy.client"},
			network:   "mainnet",
			expConfig: getDefaultDNSSyncConfig(nodeconfig.Mainnet),
		},
		{
			args:    []string{"--sync.legacy.server", "--sync.legacy.client"},
			network: "testnet",
			expConfig: func() astraconfig.DnsSync {
				cfg := getDefaultDNSSyncConfig(nodeconfig.Mainnet)
				cfg.Client = true
				cfg.Server = true
				return cfg
			}(),
		},
		{
			args:      []string{"--dns.server", "--dns.client"},
			network:   "mainnet",
			expConfig: getDefaultDNSSyncConfig(nodeconfig.Mainnet),
		},
	}

	for i, test := range tests {
		ts := newFlagTestSuite(t, dnsSyncFlags, func(command *cobra.Command, config *astraconfig.AstraConfig) {
			config.Network.NetworkType = test.network
			applyDNSSyncFlags(command, config)
		})
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.DNSSync, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.DNSSync, test.expConfig)
		}

		ts.tearDown()
	}
}
func TestSyncFlags(t *testing.T) {
	tests := []struct {
		args      []string
		network   string
		expConfig astraconfig.SyncConfig
		expErr    error
	}{
		{
			args: []string{"--sync", "--sync.downloader", "--sync.concurrency", "10", "--sync.min-peers", "10",
				"--sync.init-peers", "10", "--sync.disc.soft-low-cap", "10",
				"--sync.disc.hard-low-cap", "10", "--sync.disc.hi-cap", "10",
				"--sync.disc.batch", "10",
			},
			network: "mainnet",
			expConfig: func() astraconfig.SyncConfig {
				cfgSync := defaultMainnetSyncConfig
				cfgSync.Enabled = true
				cfgSync.Downloader = true
				cfgSync.Concurrency = 10
				cfgSync.MinPeers = 10
				cfgSync.InitStreams = 10
				cfgSync.DiscSoftLowCap = 10
				cfgSync.DiscHardLowCap = 10
				cfgSync.DiscHighCap = 10
				cfgSync.DiscBatch = 10
				return cfgSync
			}(),
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, syncFlags, func(command *cobra.Command, config *astraconfig.AstraConfig) {
			applySyncFlags(command, config)
		})
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.Sync, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Sync, test.expConfig)
		}

		ts.tearDown()
	}
}

func TestShardDataFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig astraconfig.ShardDataConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.ShardData,
		},
		{
			args: []string{"--sharddata.enable",
				"--sharddata.disk_count", "8",
				"--sharddata.shard_count", "4",
				"--sharddata.cache_time", "10",
				"--sharddata.cache_size", "512",
			},
			expConfig: astraconfig.ShardDataConfig{
				EnableShardData: true,
				DiskCount:       8,
				ShardCount:      4,
				CacheTime:       10,
				CacheSize:       512,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, shardDataFlags, func(command *cobra.Command, config *astraconfig.AstraConfig) {
			applyShardDataFlags(command, config)
		})
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.ShardData, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.ShardData, test.expConfig)
		}

		ts.tearDown()
	}
}

type flagTestSuite struct {
	t *testing.T

	cmd *cobra.Command
	hc  astraconfig.AstraConfig
}

func newFlagTestSuite(t *testing.T, flags []cli.Flag, applyFlags func(*cobra.Command, *astraconfig.AstraConfig)) *flagTestSuite {
	cli.SetParseErrorHandle(func(err error) { t.Fatal(err) })

	ts := &flagTestSuite{hc: getDefaultAstraConfigCopy(defNetworkType)}
	ts.cmd = makeTestCommand(func(cmd *cobra.Command, args []string) {
		applyFlags(cmd, &ts.hc)
	})
	if err := cli.RegisterFlags(ts.cmd, flags); err != nil {
		t.Fatal(err)
	}

	return ts
}

func (ts *flagTestSuite) run(args []string) (astraconfig.AstraConfig, error) {
	ts.cmd.SetArgs(args)
	err := ts.cmd.Execute()
	return ts.hc, err
}

func (ts *flagTestSuite) tearDown() {
	cli.SetParseErrorHandle(func(error) {})
}

func makeTestCommand(run func(cmd *cobra.Command, args []string)) *cobra.Command {
	return &cobra.Command{
		Use: "test",
		Run: run,
	}
}

func assertError(gotErr, expErr error) error {
	if (gotErr == nil) != (expErr == nil) {
		return fmt.Errorf("error unexpected [%v] / [%v]", gotErr, expErr)
	}
	if gotErr == nil {
		return nil
	}
	if !strings.Contains(gotErr.Error(), expErr.Error()) {
		return fmt.Errorf("error unexpected [%v] / [%v]", gotErr, expErr)
	}
	return nil
}

func intPtr(i int) *int {
	return &i
}
