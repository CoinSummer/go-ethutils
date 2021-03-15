package ethutils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type Chain struct {
	ChainId int64
	RpcAddr string
	Explore string
	GasPrice *big.Int
	WrapToken common.Address
	Client *ethclient.Client
}

func GetClient(rpcHost string) *ethclient.Client {
	// Connect the client
	client, err := ethclient.Dial(rpcHost)
	if err != nil {
		fmt.Printf("client connection error: " + err.Error())
		return nil
	}
	return client
}
