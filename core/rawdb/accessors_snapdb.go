package rawdb

import (
	"github.com/astra-net/astra-network/block"
	nodeconfig "github.com/astra-net/astra-network/internal/configs/node"
	"github.com/astra-net/astra-network/internal/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
)

// SnapdbInfo only used by cmd/astra/dumpdb.go
type SnapdbInfo struct {
	NetworkType         nodeconfig.NetworkType // network type
	BlockHeader         *block.Header          // latest header at snapshot
	AccountCount        uint64                 // number of dumped account
	OffchainDataDumped  bool                   // is OffchainData dumped
	IndexerDataDumped   bool                   // is IndexerData dumped
	StateDataDumped     bool                   // is StateData dumped
	DumpedSize          uint64                 // size of key-value already dumped
	LastAccountKey      hexutil.Bytes          // MPT key of the account last dumped, use this to continue dumping
	LastAccountStateKey hexutil.Bytes          // MPT key of the account's state last dumped, use this to continue dumping
}

// ReadSnapdbInfo return the SnapdbInfo of the db
func ReadSnapdbInfo(db DatabaseReader) *SnapdbInfo {
	data, _ := db.Get(snapdbInfoKey)
	if len(data) == 0 {
		return nil
	}
	info := &SnapdbInfo{}
	if err := rlp.DecodeBytes(data, info); err != nil {
		utils.Logger().Error().Err(err).Msg("Invalid SnapdbInfo RLP")
		return nil
	}
	return info
}

// WriteSnapdbInfo write the SnapdbInfo into db
func WriteSnapdbInfo(db DatabaseWriter, info *SnapdbInfo) error {
	data, err := rlp.EncodeToBytes(info)
	if err != nil {
		utils.Logger().Error().Msg("Failed to RLP encode SnapdbInfo")
		return err
	}
	if err := db.Put(snapdbInfoKey, data); err != nil {
		utils.Logger().Error().Msg("Failed to store SnapdbInfo")
		return err
	}
	return nil
}
