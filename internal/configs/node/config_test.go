package nodeconfig

import (
	"testing"

	"github.com/astra-net/astra-network/crypto/bls"

	"github.com/astra-net/astra-network/internal/blsgen"
	shardingconfig "github.com/astra-net/astra-network/internal/configs/sharding"
	"github.com/astra-net/astra-network/multibls"
	"github.com/pkg/errors"
)

func TestNodeConfigSingleton(t *testing.T) {
	// init 3 configs
	_ = GetShardConfig(2)
	// get the singleton variable
	c := GetShardConfig(Global)
	c.SetBeaconGroupID(GroupIDBeacon)
	d := GetShardConfig(Global)
	g := d.GetBeaconGroupID()
	if g != GroupIDBeacon {
		t.Errorf("GetBeaconGroupID = %v, expected = %v", g, GroupIDBeacon)
	}
}

func TestNodeConfigMultiple(t *testing.T) {
	// init 3 configs
	d := GetShardConfig(1)
	e := GetShardConfig(0)
	f := GetShardConfig(42)

	if f != nil {
		t.Errorf("expecting nil, got: %v", f)
	}

	d.SetShardGroupID("abcd")
	if d.GetShardGroupID() != "abcd" {
		t.Errorf("expecting abcd, got: %v", d.GetShardGroupID())
	}

	e.SetClientGroupID("client")
	if e.GetClientGroupID() != "client" {
		t.Errorf("expecting client, got: %v", d.GetClientGroupID())
	}
}

func TestValidateConsensusKeysForSameShard(t *testing.T) {
	// set localnet config
	networkType := "localnet"
	schedule := shardingconfig.LocalnetSchedule
	netType := NetworkType(networkType)
	SetNetworkType(netType)
	SetShardingSchedule(schedule)

	// import two keys that belong to same shard and test ValidateConsensusKeysForSameShard
	keyPath1 := "../../../.astra/1f5dbcd0504f200062ffd473f30ab1d4c2c4e01e345ca42688affda1bc66b5a37e26f6dfa3666e2f3e5dc9c8a9e30b98.key"
	priKey1, err := blsgen.LoadBLSKeyWithPassPhrase(keyPath1, "Bv1XNjrI9jE6Y0aIl3UC")
	pubKey1 := priKey1.GetPublicKey()
	if err != nil {
		t.Error(err)
	}
	keyPath2 := "../../../.astra/113c85d737e43a01dfd432b8dc40421896a82e68cd5a95b5461a3c18725ac256f5b37f65de429999d1a1b97504e1028a.key"
	priKey2, err := blsgen.LoadBLSKeyWithPassPhrase(keyPath2, "Bv1XNjrI9jE6Y0aIl3UC")
	pubKey2 := priKey2.GetPublicKey()
	if err != nil {
		t.Error(err)
	}
	keys := multibls.PublicKeys{}
	dummyKey := bls.SerializedPublicKey{}
	dummyKey.FromLibBLSPublicKey(pubKey1)
	keys = append(keys, bls.PublicKeyWrapper{Object: pubKey1, Bytes: dummyKey})
	dummyKey = bls.SerializedPublicKey{}
	dummyKey.FromLibBLSPublicKey(pubKey2)
	keys = append(keys, bls.PublicKeyWrapper{Object: pubKey2, Bytes: dummyKey})
	if err := GetDefaultConfig().ValidateConsensusKeysForSameShard(keys, 0); err != nil {
		t.Error("expected", nil, "got", err)
	}
	// add third key in different shard and test ValidateConsensusKeysForSameShard
	keyPath3 := "../../../.astra/0299970989547231d2c416f181ebed2c38407bcacc8b14c793c6d71dd6c3c0c918d9b17bce3eb6d7aeca6298f6e2ba89.key"
	priKey3, err := blsgen.LoadBLSKeyWithPassPhrase(keyPath3, "Bv1XNjrI9jE6Y0aIl3UC")
	pubKey3 := priKey3.GetPublicKey()
	if err != nil {
		t.Error(err)
	}
	dummyKey = bls.SerializedPublicKey{}
	dummyKey.FromLibBLSPublicKey(pubKey3)
	keys = append(keys, bls.PublicKeyWrapper{Object: pubKey3, Bytes: dummyKey})
	if err := GetDefaultConfig().ValidateConsensusKeysForSameShard(keys, 0); err == nil {
		e := errors.New("bls keys do not belong to the same shard")
		t.Error("expected", e, "got", nil)
	}
}
