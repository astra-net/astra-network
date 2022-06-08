package rpc

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	internal_common "github.com/astra-net/astra-network/internal/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var (
	testAddr1Str  = "0xE064a68994e9380250CfEE3E8C0e2AC5C0924548"
	testAddr2Str  = "0xb1f4fceAeF8667Ae868926c4e0817B07EDc5a938"
	testAddr1JStr = fmt.Sprintf(`"%v"`, testAddr1Str)
	testAddr2JStr = fmt.Sprintf(`"%v"`, testAddr2Str)

	testAddr1, _ = internal_common.ParseAddr(testAddr1Str)
	testAddr2, _ = internal_common.ParseAddr(testAddr2Str)
)

func TestAddressOrList_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		exp   AddressOrList
	}{
		{
			input: testAddr1JStr,
			exp: AddressOrList{
				Address:     &testAddr1,
				AddressList: nil,
			},
		},
		{
			input: fmt.Sprintf("[%v, %v]", testAddr1JStr, testAddr2JStr),
			exp: AddressOrList{
				Address:     nil,
				AddressList: []common.Address{testAddr1, testAddr2},
			},
		},
	}

	for _, test := range tests {
		var aol *AddressOrList
		if err := json.Unmarshal([]byte(test.input), &aol); err != nil {
			t.Fatal(err)
		}
		if err := checkAddressOrListEqual(aol, &test.exp); err != nil {
			t.Error(err)
		}
	}
}

func checkAddressOrListEqual(a, b *AddressOrList) error {
	if (a.Address != nil) != (b.Address != nil) {
		return errors.New("address not equal")
	}
	if a.Address != nil && *a.Address != *b.Address {
		return errors.New("address not equal")
	}
	if len(a.AddressList) != len(b.AddressList) {
		return errors.New("address list size not equal")
	}
	for i, addr1 := range a.AddressList {
		addr2 := b.AddressList[i]
		if addr1 != addr2 {
			return errors.New("address list elem not equal")
		}
	}
	return nil
}

func TestDelegation_IntoStructuredResponse(t *testing.T) {
	d := Delegation{
		ValidatorAddress: "0xbBb3A39023B5c3480014e7A1b2F01C3c8C23f92d",
		DelegatorAddress: "0xc5092ED2b686308dcEB390A8EAE00493ae0eAFFf",
		Amount:           big.NewInt(1000),
		Reward:           big.NewInt(1014),
		Undelegations:    make([]Undelegation, 0),
	}
	rs1, err := NewStructuredResponse(d)
	require.NoError(t, err)

	rs2 := d.IntoStructuredResponse()

	js1, err := json.Marshal(rs1)
	require.NoError(t, err)

	js2, err := json.Marshal(rs2)
	require.NoError(t, err)

	require.JSONEq(t, string(js1), string(js2))
}
