package key_manager

import (
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"

	crypto "github.com/ethereum/go-ethereum/crypto"
)

// Purpose BIP43 - Purpose Field for Deterministic Wallets
// https://github.com/bitcoin/bips/blob/master/bip-0043.mediawiki
//
// Purpose is a constant set to 44' (or 0x8000002C) following the BIP43 recommendation.
// It indicates that the subtree of this node is used according to this specification.
//
// What does 44' mean in BIP44?
// https://bitcoin.stackexchange.com/questions/74368/what-does-44-mean-in-bip44
//
// 44' means that hardened keys should be used. The distinguisher for whether
// a key a given index is hardened is that the index is greater than 2^31,
// which is 2147483648. In hex, that is 0x80000000. That is what the apostrophe (') means.
// The 44 comes from adding it to 2^31 to get the final hardened key index.
// In hex, 44 is 2C, so 0x80000000 + 0x2C = 0x8000002C.
type Purpose = uint32

const (
	PurposeBIP44 Purpose = 0x8000002C // 44' BIP44
	PurposeBIP49 Purpose = 0x80000031 // 49' BIP49
	PurposeBIP84 Purpose = 0x80000054 // 84' BIP84
)

// CoinType SLIP-0044 : Registered coin types for BIP-0044
// https://github.com/satoshilabs/slips/blob/master/slip-0044.md
type CoinType = uint32

const (
	CoinTypeBTC CoinType = 0x80000000
	CoinTypeLTC CoinType = 0x80000002
	CoinTypeETH CoinType = 0x8000003c
	CoinTypeEOS CoinType = 0x800000c2
)

const (
	Apostrophe uint32 = 0x80000000 // 0'
)

type Key struct {
	path     string
	bip32Key *bip32.Key
}

func (k *Key) Encode(compress bool) (wif, address, segwitBech32, segwitNested string, privateKey string, err error) {
	prvKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), k.bip32Key.Key)
	return GenerateFromBytes(prvKey, compress)
}

func (k *Key) EncodeEth() (address string, publicKey string, privateKey string) {
	prvKey, pubKey := btcec.PrivKeyFromBytes(btcec.S256(), k.bip32Key.Key)
	pub := pubKey.ToECDSA()
	address = crypto.PubkeyToAddress(*pub).Hex()
	publicKey = string(crypto.FromECDSAPub(pub))
	privateKey = fmt.Sprintf("%x", prvKey.D)
	for i := len(privateKey); i < 64; i++ {
		privateKey = "0" + privateKey
	}
	return address, publicKey, privateKey
}

// https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
// bip44 define the following 5 levels in BIP32 path:
// m / purpose' / coin_type' / account' / change / address_index

func (k *Key) GetPath() string {
	return k.path
}

type KeyManager struct {
	mnemonic   string
	passphrase string
	keys       map[string]*bip32.Key
	mux        sync.Mutex
}

// NewKeyManager return new key manager
// bitSize has to be a multiple 32 and be within the inclusive range of {128, 256}
// 128: 12 phrases
// 256: 24 phrases
func NewKeyManager(bitSize int, passphrase string) (*KeyManager, error) {
	entropy, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	km := &KeyManager{
		mnemonic:   mnemonic,
		passphrase: passphrase,
		keys:       make(map[string]*bip32.Key, 0),
	}
	return km, nil
}

// NewKeyManager return new key manager
// bitSize has to be a multiple 32 and be within the inclusive range of {128, 256}
// 128: 12 phrases
// 256: 24 phrases
func NewKeyManagerWithMnemonic(bitSize int, passphrase string, keyphrase string) (*KeyManager, error) {

	entropy, err := bip39.EntropyFromMnemonic(keyphrase)
	if err != nil {
		return nil, err
	}

	/*
		valid := bip39.IsMnemonicValid(keyphrase)
		if valid == false {
			err := errors.New("mnemonic not valid")
			return nil, err
		}
	*/

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	km := &KeyManager{
		mnemonic:   mnemonic,
		passphrase: passphrase,
		keys:       make(map[string]*bip32.Key, 0),
	}
	return km, nil
}

func (km *KeyManager) GetMnemonic() string {
	return km.mnemonic
}

func (km *KeyManager) GetPassphrase() string {
	return km.passphrase
}

func (km *KeyManager) GetSeed() []byte {
	return bip39.NewSeed(km.GetMnemonic(), km.GetPassphrase())
}

func (km *KeyManager) getKey(path string) (*bip32.Key, bool) {
	km.mux.Lock()
	defer km.mux.Unlock()

	key, ok := km.keys[path]
	return key, ok
}

func (km *KeyManager) setKey(path string, key *bip32.Key) {
	km.mux.Lock()
	defer km.mux.Unlock()

	km.keys[path] = key
}

func (km *KeyManager) GetMasterKey() (*bip32.Key, error) {
	path := "m"

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	key, err := bip32.NewMasterKey(km.GetSeed())
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetPurposeKey(purpose uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'`, purpose-Apostrophe)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetMasterKey()
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(purpose)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetCoinTypeKey(purpose, coinType uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'`, purpose-Apostrophe, coinType-Apostrophe)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetPurposeKey(purpose)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(coinType)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetAccountKey(purpose, coinType, account uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'`, purpose-Apostrophe, coinType-Apostrophe, account)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetCoinTypeKey(purpose, coinType)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(account + Apostrophe)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

// GetChangeKey ...
// https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki#change
// change constant 0 is used for external chain
// change constant 1 is used for internal chain (also known as change addresses)
func (km *KeyManager) GetChangeKey(purpose, coinType, account, change uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d`, purpose-Apostrophe, coinType-Apostrophe, account, change)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	parent, err := km.GetAccountKey(purpose, coinType, account)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(change)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetShortAccountKey(purpose uint32, coinType uint32, account uint32) (*bip32.Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d`, purpose-Apostrophe, coinType-Apostrophe, account)

	key, ok := km.getKey(path)
	if ok {
		return key, nil
	}

	key, err := km.GetAccountKey(purpose, coinType, account)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return key, nil
}

func (km *KeyManager) GetShortKey(purpose uint32, coinType uint32, account uint32, index uint32) (*Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d`, purpose-Apostrophe, coinType-Apostrophe, account, index)

	key, ok := km.getKey(path)
	if ok {
		return &Key{path: path, bip32Key: key}, nil
	}

	parent, err := km.GetShortAccountKey(purpose, coinType, account)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(index)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return &Key{path: path, bip32Key: key}, nil
}

func (km *KeyManager) GetKey(purpose, coinType, account, change, index uint32) (*Key, error) {
	path := fmt.Sprintf(`m/%d'/%d'/%d'/%d/%d`, purpose-Apostrophe, coinType-Apostrophe, account, change, index)

	key, ok := km.getKey(path)
	if ok {
		return &Key{path: path, bip32Key: key}, nil
	}

	parent, err := km.GetChangeKey(purpose, coinType, account, change)
	if err != nil {
		return nil, err
	}

	key, err = parent.NewChildKey(index)
	if err != nil {
		return nil, err
	}

	km.setKey(path, key)

	return &Key{path: path, bip32Key: key}, nil
}

func Generate(compress bool) (wif, address, segwitBech32, segwitNested string, privateKey string, err error) {
	prvKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return "", "", "", "", "", err
	}
	return GenerateFromBytes(prvKey, compress)
}

func GenerateFromBytes(prvKey *btcec.PrivateKey, compress bool) (wif, address, segwitBech32, segwitNested string, privateKey string, err error) {
	// generate the wif(wallet import format) string
	btcwif, err := btcutil.NewWIF(prvKey, &chaincfg.MainNetParams, compress)
	if err != nil {
		return "", "", "", "", "", err
	}
	wif = btcwif.String()

	// generate a normal p2pkh address
	serializedPubKey := btcwif.SerializePubKey()
	addressPubKey, err := btcutil.NewAddressPubKey(serializedPubKey, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", "", err
	}
	address = addressPubKey.EncodeAddress()

	// generate a normal p2wkh address from the pubkey hash
	witnessProg := btcutil.Hash160(serializedPubKey)
	addressWitnessPubKeyHash, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", "", err
	}
	segwitBech32 = addressWitnessPubKeyHash.EncodeAddress()

	// generate an address which is
	// backwards compatible to Bitcoin nodes running 0.6.0 onwards, but
	// allows us to take advantage of segwit's scripting improvments,
	// and malleability fixes.
	serializedScript, err := txscript.PayToAddrScript(addressWitnessPubKeyHash)
	if err != nil {
		return "", "", "", "", "", err
	}
	addressScriptHash, err := btcutil.NewAddressScriptHash(serializedScript, &chaincfg.MainNetParams)
	if err != nil {
		return "", "", "", "", "", err
	}
	segwitNested = addressScriptHash.EncodeAddress()

	privateKey = fmt.Sprintf("%x", prvKey.D)

	return wif, address, segwitBech32, segwitNested, privateKey, nil
}
