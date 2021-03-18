package ethutils

import (
	"github.com/deng00/go-base/config"
	"github.com/deng00/go-base/logging"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"strconv"
)

type Config struct {
	config *config.Config
	logger *logging.SugaredLogger
}

func NewConfig(config *config.Config, logger *logging.SugaredLogger) *Config {
	return &Config{
		config: config,
		logger: logger,
	}
}

func (c *Config) GetChainInfo() *Chain {
	chainName := c.config.GetString("chain")
	return ParserChainInfo(c.config.GetStringMapString("chains." + chainName))
}

func (c *Config) MustGetAccount() *Account {
	accountName := c.config.GetString("account")
	if accountName == "" {
		c.logger.Fatalf("account not set")
		os.Exit(1)
	}
	account := c.GetAccount(accountName)
	if account == nil {
		c.logger.Fatalf("account %s not set", accountName)
		os.Exit(1)
	}
	return account
}

func (c *Config) MustGetToken(token string) Token {
	IsSet := c.config.IsSet("tokens." + token)
	if !IsSet {
		c.logger.Fatalf("token %s not set", token)
		os.Exit(1)
	}
	tokenInfo := c.config.GetStringMapString("tokens." + token)
	_decimal, _ := strconv.Atoi(tokenInfo["decimal"])
	address := HexToAddress(tokenInfo["address"])
	return Token{
		Address: &address,
		Decimal: _decimal,
	}
}

func (c *Config) GetAccount(accountName string) *Account {
	pStr := c.config.GetString("accounts." + accountName + ".key")
	if pStr == "" {
		return nil
	}
	var _account *Account
	if !IsMnemonic(pStr) {
		_account = GetAccountFromPStr(pStr)
	} else {
		_account = GetAccountFromMnemonic(pStr, 0)
	}
	_account.Client = c.GetChainInfo().Client
	c.logger.Infof("using account %s", _account.Address.Hex())
	return _account
}

func (c *Config) MustGetHDAccountByIndex(_index int) *Account {
	accountName := c.config.GetString("account")
	if accountName == "" {
		c.logger.Fatalf("account not set")
		os.Exit(1)
	}
	pStr := c.config.GetString("accounts." + accountName + ".key")
	_account := GetAccountFromMnemonic(pStr, _index)
	if _account == nil {
		c.logger.Infof("account parse error")
		os.Exit(1)
	}
	return _account
}

func (c *Config) MustGetContractAddress(name string) *common.Address {
	cStr := c.config.GetString("contracts." + name)
	if cStr == "" {
		c.logger.Fatalf("contract %s not found", name)
		os.Exit(1)
	}
	contract := HexToAddress(cStr)
	return &contract
}
