package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	ethermint "github.com/tharsis/ethermint/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/evmos/app"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

type ProposalTestSuite struct {
	suite.Suite

	ctx          sdk.Context
	app          *app.Evmos
	queryClient  types.QueryClient
	dynamicTxFee bool

	address     common.Address
	consAddress sdk.ConsAddress
	// for generate test tx
	clientCtx client.Context
	ethSigner ethtypes.Signer
	// appCodec  codec.Codec
	signer keyring.Signer
}

func (suite *ProposalTestSuite) DoSetupTest(t require.TestingT) {
	checkTx := false

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = tests.NewSigner(priv)

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	if suite.dynamicTxFee {
		// setup feemarketGenesis params
		feemarketGenesis := feemarkettypes.DefaultGenesisState()
		feemarketGenesis.Params.EnableHeight = 1
		feemarketGenesis.Params.NoBaseFee = false
		feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)
		suite.app = app.Setup(checkTx, feemarketGenesis)
	} else {
		suite.app = app.Setup(checkTx, nil)
	}

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "evmos_9000-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),

		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})
	suite.app.EvmKeeper.WithContext(suite.ctx)

	// queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	// types.RegisterQueryServer(queryHelper, suite.app.CronKeeper)
	// suite.queryClient = types.NewQueryClient(queryHelper)

	// // Query helper from ethermint to query the denom
	// queryHelperEvm := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	// evmtypes.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	// suite.queryClientEvm = evmtypes.NewQueryClient(queryHelperEvm)

	acc := &ethermint.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	// Set the sequence as 1 because that's the value that
	// the acc will have after the antehandler executes
	account := suite.app.AccountKeeper.GetAccount(suite.ctx, suite.address.Bytes())
	account.SetSequence(1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, account)

	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
	// suite.appCodec = encodingConfig.Marshaler

	suite.Require().NoError(err)
}

func (suite *ProposalTestSuite) SetupTest() {
	suite.DoSetupTest(suite.T())
}

func (suite *ProposalTestSuite) TestRegisterTokenPair() {
	suite.SetupTest()
	pair := types.NewTokenPair(tests.GenerateAddress(), "ramacoin", true)
	err := suite.app.IntrarelayerKeeper.RegisterTokenPair(suite.ctx, pair)
	suite.Require().NoError(err)
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalTestSuite))
}
