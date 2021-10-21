package contracts

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	//go:embed ERC20Burnable.json
	compiledContractJSON string
	ERC20ABI             abi.ABI
)

func NewErc20Contract() (abi.ABI, error) {
	return abi.JSON(strings.NewReader(compiledContractJSON))
}
