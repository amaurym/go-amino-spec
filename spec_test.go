package amino_spec

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
)

func TestAmino(t *testing.T) {
	testcases := []struct {
		name    string
		msg     proto.Message
		expJSON string
	}{
		// with vs without omitempty
		{
			"with omitempty (default bahavior)",
			&group.MsgVote{ProposalId: 0}, // ProposalId has omitempty
			`{"type":"cosmos-sdk/group/MsgVote","value":{}}`,
		},
		{
			"jsontag without omitempty",
			&govv1.MsgVote{ProposalId: 0}, // ProposalId doesn't have omitempty
			`{"type":"cosmos-sdk/v1/MsgVote","value":{"proposal_id":"0"}}`,
		},
		// pubkeys
		{
			"secp256k1 (only inner key field)",
			&secp256k1.PubKey{Key: []byte{1}},
			`{"type":"tendermint/PubKeySecp256k1","value":"AQ=="}`, // value is not {"key":"AQ=="}
		},
		{
			"ed25519 (only inner key field)",
			&ed25519.PubKey{Key: []byte{1}},
			`{"type":"tendermint/PubKeyEd25519","value":"AQ=="}`, // value is not {"key":"AQ=="}
		},
		{
			"multisig (threshold is string)",
			&multisig.LegacyAminoPubKey{Threshold: 2},
			`{"type":"tendermint/PubKeyMultisigThreshold","value":{"pubkeys":[],"threshold":"2"}}`, // value is not {"pubkeys":[],"threshold":2}
		},
		// gogoproto.nullable= true vs false
		// it seems that whenever we add gogoproto.nullable=false, then we don't have omitempty either
		{
			"nullable=false",
			&govv1beta1.GenesisState{}, // TallyParams has proto annotation nullable=false (golang: struct)
			`{"deposit_params":{},"deposits":null,"proposals":null,"tally_params":{},"votes":null,"voting_params":{}}`,
		},
		{
			"nullable=true",
			&govv1.GenesisState{}, // TallyParams doesn't have any proto annotation (golang: pointer to a struct)
			`{}`,
		},
		// gogoproto.nullable=false, with or without explicit omitempty
		{
			"nullable=false, omitempty=true",
			&govv1.DepositParams{}, // MinDeposit has nullable=false (golang: pointer) and jsontag omitempty=true
			`{}`,
		},
		{
			"nullable=false, omitempty=false",
			&govv1.Deposit{}, // Amount has nullable=false (golang: pointer) and jsontag omitempty=false
			`{"amount":null}`,
		},
	}

	cdc := simapp.MakeTestEncodingConfig()
	group.RegisterLegacyAminoCodec(cdc.Amino)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := cdc.Amino.MarshalJSON(tc.msg)
			require.NoError(t, err)
			out, err = sdk.SortJSON(out)
			require.NoError(t, err)

			require.Equal(t, tc.expJSON, string(out))
		})
	}
}
