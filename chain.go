package ethutils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strconv"
)

type Chain struct {
	ChainId       int64
	RpcAddr       string
	Explore       string
	GasPrice      *big.Int
	WrapToken     common.Address
	BatchContract common.Address
	Client        *ethclient.Client
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

func ParserChainInfo(chainInfoMap map[string]string) *Chain {
	// TODO: check key
	_chainId, _ := strconv.Atoi(chainInfoMap["chain_id"])
	_gasPrice, _ := strconv.Atoi(chainInfoMap["gas_price"])
	chainInfo := &Chain{
		ChainId:       int64(_chainId),
		RpcAddr:       chainInfoMap["rpc_addr"],
		Explore:       chainInfoMap["explore"],
		GasPrice:      big.NewInt(int64(_gasPrice) * GWEI),
		WrapToken:     HexToAddress(chainInfoMap["wrap_token"]),
		BatchContract: HexToAddress(chainInfoMap["batch_contract"]),
	}
	chainInfo.Client = GetClient(chainInfo.RpcAddr)
	return chainInfo
}
