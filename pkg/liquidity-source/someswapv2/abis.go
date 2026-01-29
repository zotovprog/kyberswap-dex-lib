package someswapv2

import (
	"bytes"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var factoryABI abi.ABI

func FactoryABI() abi.ABI {
	return factoryABI
}

func init() {
	var err error
	factoryABI, err = abi.JSON(bytes.NewReader(factoryABIJson))
	if err != nil {
		panic(err)
	}
}

