package ethutils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"strconv"
)

type Chain struct {
	ChainId       int64
	RpcAddr       string
	WsAddr        string
	Explore       string
	GasPrice      *big.Int
	WrapToken     common.Address
	BatchContract common.Address
	Client        *ethclient.Client
	WsClient      *rpc.Client
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

func GetWsClient(rpcHost string) *rpc.Client {
	// Connect the client
	if rpcHost == "" {
		return nil
	}
	client, err := rpc.Dial(rpcHost)
	if err != nil {
		fmt.Printf("client connection error: " + err.Error())
		return nil
	}
	return client
}

func ParserChainInfo(chainInfoMap map[string]string) *Chain {
	// TODO: check key
	_chainId, _ := strconv.Atoi(chainInfoMap["chain_id"])
	_gasPrice, _ := strconv.ParseFloat(chainInfoMap["gas_price"], 64)
	chainInfo := &Chain{
		ChainId:       int64(_chainId),
		RpcAddr:       chainInfoMap["rpc_addr"],
		WsAddr:        chainInfoMap["ws_addr"],
		Explore:       chainInfoMap["explore"],
		GasPrice:      big.NewInt(int64(_gasPrice * GWEI)),
		WrapToken:     HexToAddress(chainInfoMap["wrap_token"]),
		BatchContract: HexToAddress(chainInfoMap["batch_contract"]),
	}
	chainInfo.Client = GetClient(chainInfo.RpcAddr)
	chainInfo.WsClient = GetWsClient(chainInfo.WsAddr)
	return chainInfo
}
