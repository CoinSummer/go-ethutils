package ethutils

import "github.com/ethereum/go-ethereum/common"

type Token struct {
	Address *common.Address
	Decimal int
	Symbol  string
	Name    string
}
