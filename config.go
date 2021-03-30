package ethutils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/howeyc/gopass"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type ConfigInterface interface {
	IsSet(key string) bool
	GetInt(key string) int
	GetString(key string) string
	GetStringMapString(key string) map[string]string
}

type Config struct {
	config ConfigInterface
}

func NewConfig(config ConfigInterface) *Config {
	return &Config{
		config: config,
	}
}

func (c *Config) MustGetChainInfo() *Chain {
	chainName := c.config.GetString("chain")
	if chainName == "" {
		fmt.Printf("chain not set")
		os.Exit(1)
	}
	return c.GetChainInfo(c.config.GetString("chain"))
}

func (c *Config) GetChainInfo(chainName string) *Chain {
	return ParserChainInfo(c.config.GetStringMapString("chains." + chainName))
}

func (c *Config) MustGetAccount() *Account {
	accountName := c.config.GetString("account")
	if accountName == "" {
		fmt.Printf("account not set")
		os.Exit(1)
	}
	account := c.GetAccount(accountName)
	if account == nil {
		fmt.Printf("account %s not set", accountName)
		os.Exit(1)
	}
	return account
}

func getPassLoop(keystoreBytes []byte) string {
	fmt.Println("please enter your keystore file password:")
	for {
		password, _ := gopass.GetPasswdMasked()
		_, err := keystore.DecryptKey(keystoreBytes, string(password))
		if err == nil {
			fmt.Println("decrypt success")
			return string(password)
		}
		fmt.Println(err.Error())
		fmt.Println("please reenter it:")
	}
}

func (c *Config) GetAccount(accountName string) *Account {
	pStr := c.config.GetString("accounts." + accountName + ".key")
	if pStr == "" {
		return nil
	}
	var _account *Account
	if IsMnemonic(pStr) {
		_account = GetAccountFromMnemonic(pStr, c.config.GetInt("account_index"))
	} else if strings.Contains(pStr, "/") {
		keystoreBytes, err := ioutil.ReadFile(pStr)
		if err != nil {
			fmt.Printf("can not read keystore file content")
			return nil
		}
		// keystore file
		password := getPassLoop(keystoreBytes)
		_account = GetAccountFromKS(pStr, password)
	} else {
		_account = GetAccountFromPStr(pStr)
	}
	if _account == nil {
		return nil
	}
	_account.Client = c.MustGetChainInfo().Client
	return _account
}

func (c *Config) MustGetToken(token string) *Token {
	IsSet := c.config.IsSet("tokens." + token)
	if !IsSet {
		fmt.Printf("token %s not set", token)
		os.Exit(1)
	}
	return c.GetToken(token)
}

func (c *Config) GetToken(token string) *Token {
	tokenInfo := c.config.GetStringMapString("tokens." + token)
	_decimal, _ := strconv.Atoi(tokenInfo["decimal"])
	address := HexToAddress(tokenInfo["address"])
	return &Token{
		Address: &address,
		Decimal: _decimal,
	}
}

func (c *Config) MustGetHDAccountByIndex(_index int) *Account {
	accountName := c.config.GetString("account")
	if accountName == "" {
		fmt.Printf("account not set")
		os.Exit(1)
	}
	pStr := c.config.GetString("accounts." + accountName + ".key")
	_account := GetAccountFromMnemonic(pStr, _index)
	if _account == nil {
		fmt.Printf("account %s parse error", accountName)
		os.Exit(1)
	}
	return _account
}

func (c *Config) MustGetContractAddress(name string) *common.Address {
	cStr := c.config.GetString("contracts." + name)
	if cStr == "" {
		fmt.Printf("contract %s not found", name)
		os.Exit(1)
	}
	contract := HexToAddress(cStr)
	return &contract
}
