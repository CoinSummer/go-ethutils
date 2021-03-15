package ethutils

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

func MakeTxOpts(account *Account, value *big.Int, gasPrice *big.Int, gasLimit uint64, chainID int64) *bind.TransactOpts {
	txOpts := &bind.TransactOpts{
		From:  account.Address,
		Nonce: account.GetNonce(),
		Signer: func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			var txSigner types.Signer
			if chainID != 0 {
				// EIP155 signer
				txSigner = types.NewEIP155Signer(big.NewInt(chainID))
			} else {
				// default is homestead signer
				txSigner = types.HomesteadSigner{}
			}
			signedTx, err := types.SignTx(tx, txSigner, account.PrivateKey)
			if err != nil {
				return nil, err
			}
			return signedTx, nil
		},
		Value:    value,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}
	return txOpts
}

func MakeCallOpts(from common.Address) *bind.CallOpts {
	return &bind.CallOpts{
		Pending:     false,
		From:        from,
		BlockNumber: nil,
		Context:     context.Background(),
	}
}
