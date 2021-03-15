package ethutils

import (
	"context"
	"crypto/ecdsa"
	"github.com/deng00/ethutils/key_manager"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/tools/go/ssa/interp/testdata/src/fmt"
	"io/ioutil"
	"math/big"
	"strings"
)

func HexToAddress(s string) common.Address {
	return common.HexToAddress(s)
}

func StrToPri(pkStr string) *ecdsa.PrivateKey {
	privateKey, err := crypto.HexToECDSA(pkStr)
	if err != nil {
		return nil
	}
	return privateKey
}

type Account struct {
	Client     *ethclient.Client
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
	nonce      uint64
}

func (a *Account) SetNonce(nonce uint64) {
	a.nonce = nonce
}

func (a *Account) GetNonce() *big.Int {
	value := a.nonce
	if value == 0 {
		nonce, err := a.Client.PendingNonceAt(context.Background(), a.Address)
		if err != nil {
			return nil
		}
		value = nonce
	}
	a.nonce = value + 1
	return big.NewInt(int64(value))
}

func (a *Account) GetNonceUint64() uint64 {
	return a.GetNonce().Uint64()
}

func GetAccountFromPStr(pkStr string) *Account {
	priKey := StrToPri(pkStr)
	if priKey == nil {
		return nil
	}
	return &Account{
		Address:    crypto.PubkeyToAddress(priKey.PublicKey),
		PrivateKey: priKey,
	}
}

func GetAccountFromMnemonic(mnemonic string, index int) *Account {
	km, _ := key_manager.NewKeyManagerWithMnemonic(256, "", mnemonic)
	key, err := km.GetKey(key_manager.PurposeBIP44, key_manager.CoinTypeETH, 0, 0, uint32(index))
	if err != nil {
		fmt.Printf(err)
		return nil
	}
	address, _, privateKey := key.EncodeEth()
	priKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		fmt.Printf("convent private key error: %s", err.Error())
		return nil
	}
	return &Account{
		Address:    HexToAddress(strings.ToLower(address)),
		PrivateKey: priKey,
	}
}

func GetAccountFromKS(fromKeyStoreFile, password string) *Account {
	fromKeystore, err := ioutil.ReadFile(fromKeyStoreFile)
	if err != nil {
		return nil
	}
	fromKey, err := keystore.DecryptKey(fromKeystore, password)
	if err != nil {
		return nil
	}
	fromPrivateKey := fromKey.PrivateKey
	fromAddr := crypto.PubkeyToAddress(fromPrivateKey.PublicKey)
	return &Account{
		Address:    HexToAddress(strings.ToLower(fromAddr.Hex())),
		PrivateKey: fromPrivateKey,
	}
}

func GetBalance(client *ethclient.Client, address common.Address) float64 {
	balance, err := client.PendingBalanceAt(context.Background(), address)
	if err != nil {
		return 0
	}
	return FormatAmount(balance, 18)
}

func NewMnemonic(bitSize int, passphrase string) (*key_manager.KeyManager, error) {
	return key_manager.NewKeyManager(bitSize, passphrase)
}

func PrivateKeyToHex(pk *ecdsa.PrivateKey) string {
	privateKeyBytes := crypto.FromECDSA(pk)
	return hexutil.Encode(privateKeyBytes)[2:]
}
