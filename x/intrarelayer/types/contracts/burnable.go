package contracts

import (
	_ "embed"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tharsis/evmos/x/intrarelayer/types"
)

var (
	//go:embed ERC20Burnable.json
	ERC20BurnableJSON []byte

	// ModuleCRC20Contract is the compiled cronos erc20 contract
	ERC20BurnableContract CompiledContract

	// EVMModuleAddress is the native module address for EVM
	ERC20BurnableAddress common.Address
)

func init() {
	ERC20BurnableAddress = types.ModuleAddress

	err := json.Unmarshal(ERC20BurnableJSON, &ERC20BurnableContract)
	if err != nil {
		panic(err)
	}
}
