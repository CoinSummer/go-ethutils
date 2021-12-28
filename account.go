package ethutils

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/deng00/ethutils/key_manager"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"math/big"
	"strconv"
	"strings"
)

var ZeroAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")

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
	km, err := key_manager.NewKeyManagerWithMnemonic(256, "", mnemonic)
	if err != nil {
		fmt.Printf("mnemonic error")
		return nil
	}
	key, err := km.GetKey(key_manager.PurposeBIP44, key_manager.CoinTypeETH, 0, 0, uint32(index))
	if err != nil {
		fmt.Printf("mnemonic error")
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

func GetAccountFromKS(keystoreFilePath, password string) *Account {
	keystoreBytes, err := ioutil.ReadFile(keystoreFilePath)
	if err != nil {
		return nil
	}
	key, err := keystore.DecryptKey(keystoreBytes, password)
	if err != nil {
		return nil
	}
	fromAddr := crypto.PubkeyToAddress(key.PrivateKey.PublicKey)
	return &Account{
		Address:    HexToAddress(strings.ToLower(fromAddr.Hex())),
		PrivateKey: key.PrivateKey,
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

func IsMnemonic(words string) bool {
	l := len(strings.Split(words, " "))
	return l == 12 || l == 24
}

func PrivateKeyToHex(pk *ecdsa.PrivateKey) string {
	privateKeyBytes := crypto.FromECDSA(pk)
	return hexutil.Encode(privateKeyBytes)[2:]
}

func Sign(msg []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	ethMessage := append([]byte("\x19Ethereum Signed Message:\n"+strconv.Itoa(len(msg))), msg...)
	hash := crypto.Keccak256(ethMessage)
	return crypto.Sign(hash, key)
}
