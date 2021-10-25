package contracts

import (
	_ "embed"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

var (
	//go:embed ERC20PresentMinterPauser.json
	ERC20BurnableAndMintableJSON []byte

	// ModuleCRC20Contract is the compiled cronos erc20 contract
	ERC20BurnableAndMintableContract CompiledContract

	// EVMModuleAddress is the native module address for EVM
	ERC20BurnableAndMintableAddress common.Address
)

func init() {
	ERC20BurnableAndMintableAddress = types.ModuleAddress

	err := json.Unmarshal(ERC20BurnableAndMintableJSON, &ERC20BurnableAndMintableContract)
	if err != nil {
		panic(err)
	}

	if len(ERC20BurnableAndMintableContract.Bin) == 0 {
		panic("load contract failed")
	}
}
