package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	"github.com/tharsis/evmos/x/intrarelayer/types"
	"github.com/tharsis/evmos/x/intrarelayer/types/contracts"
)

// RegisterTokenPair registers token pair by coin denom and ERC20 contract
// address. This function fails if the mapping ERC20 <--> cosmos coin already exists.
func (k Keeper) RegisterTokenPair(ctx sdk.Context, pair types.TokenPair) error {
	params := k.GetParams(ctx)
	if !params.EnableIntrarelayer {
		return sdkerrors.Wrap(types.ErrInternalTokenPair, "intrarelaying is currently disabled by governance")
	}

	erc20 := pair.GetERC20Contract()
	if k.IsERC20Registered(ctx, erc20) {
		return sdkerrors.Wrapf(types.ErrInternalTokenPair, "token ERC20 contract already registered: %s", pair.Erc20Address)
	}

	if k.IsDenomRegistered(ctx, pair.Denom) {
		return sdkerrors.Wrapf(types.ErrInternalTokenPair, "coin denomination already registered: %s", pair.Denom)
	}

	// create metadata if not already stored
	if err := k.CreateMetadata(ctx, pair); err != nil {
		return sdkerrors.Wrap(err, "failed to create wrapped coin denom metadata for ERC20")
	}

	k.SetTokenPair(ctx, pair)
	return nil
}

func (k Keeper) CreateMetadata(ctx sdk.Context, bridge types.TokenPair) error {
	// TODO: replace for HasDenomMetaData once available
	_, found := k.bankKeeper.GetDenomMetaData(ctx, bridge.Denom)
	if found {
		// metadata already exists; exit
		// TODO: validate that the fields from the ERC20 match the denom metadata's
		return nil
	}

	// if cosmos denom doesn't exist
	// TODO: query the contract and supply
	erc20, err := contracts.NewErc20Contract()
	ctorArgs, err := erc20.Pack("symbol")
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to create ABI for erc20: %s", err.Error())
	}
	// data := append(contracts.ERC20BurnableContract.Bin, ctorArgs...)
	// if err != nil {
	// return sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to create the bin ABI for erc20: %s", err.Error())
	// }

	encoded_msg := (hexutil.Bytes)(hexutil.Encode(ctorArgs))

	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to create ABI for erc20: %s", err.Error())
	}
	// &evmtypes.TransactionArgs{
	// 	From: &suite.address,
	// 	Data: (*hexutil.Bytes)(&data),
	// }

	from := common.Address{}
	to := common.HexToAddress(bridge.Erc20Address)
	gas := hexutil.Uint64(0x5208)
	args := &evmtypes.TransactionArgs{
		From: &from,
		To:   &to,
		Gas:  &gas, //hexutils.HexToBytes("0x5208"),hexutil.Uint64(20000)
		// GasPrice: "0x55ae82600",
		// Value:    "0x16345785d8a0000",
		Data: &encoded_msg, //"0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675",
	}

	bz, err := json.Marshal(&args)
	if err != nil {
		return err
	}

	// baseFee, err := e.BaseFee()
	// if err != nil {
	// 	return 0, err
	// }

	// var bf *sdk.Int
	// if baseFee != nil {
	// 	aux := sdk.NewIntFromBigInt(baseFee)
	// 	bf = &aux
	// }

	req := evmtypes.EthCallRequest{
		Args:   bz,
		GasCap: 100000,
	}

	msg, err := k.evmKeeper.EthCall(ctx.Context(), &req)
	_ = msg
	// a := erc20.getArguments("symbol",[])

	symbol := "rama"
	decimals := uint32(18)
	token := "rama"

	// create a bank denom metadata based on the ERC20 token ABI details
	metadata := banktypes.Metadata{
		Description: fmt.Sprintf("Cosmos coin token wrapper of %s ", token),
		Base:        bridge.Denom,
		// NOTE: Denom units MUST be increasing
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    bridge.Denom,
				Exponent: 0,
			},
			{
				Denom:    token,
				Exponent: decimals,
			},
		},
		Name:   token,
		Symbol: symbol,
	}

	if err := metadata.Validate(); err != nil {
		return sdkerrors.Wrapf(err, "ERC20 token data is invalid for contract %s", bridge.Erc20Address)
	}

	k.bankKeeper.SetDenomMetaData(ctx, metadata)
	return nil
}

// EnableRelay enables relaying for a given token pair
func (k Keeper) EnableRelay(ctx sdk.Context, token string) (types.TokenPair, error) {
	id := k.GetTokenPairID(ctx, token)

	if len(id) == 0 {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "token %s not registered", token)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "not registered")
	}

	pair.Enabled = true

	k.SetTokenPair(ctx, pair)
	return pair, nil
}

// UpdateTokenPairERC20 updates the ERC20 token address for the registered token pair
func (k Keeper) UpdateTokenPairERC20(ctx sdk.Context, erc20Addr, newERC20Addr common.Address) (types.TokenPair, error) {
	id := k.GetERC20Map(ctx, erc20Addr)
	if len(id) == 0 {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "token %s not registered", erc20Addr)
	}

	pair, found := k.GetTokenPair(ctx, id)
	if !found {
		return types.TokenPair{}, sdkerrors.Wrapf(types.ErrInternalTokenPair, "not registered")
	}

	pair.Erc20Address = newERC20Addr.Hex()
	k.SetTokenPair(ctx, pair)
	return pair, nil
}
