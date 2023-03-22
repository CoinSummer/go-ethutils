package go_ethutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// 非对称加密算法，增强数据传输中的安全性

type ByteSlice []byte

func (m ByteSlice) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(m))
}

type EncryptOption struct {
	Iv             ByteSlice `json:"iv"`
	EphemPublicKey ByteSlice `json:"ephemPublicKey"`
	Ciphertext     ByteSlice `json:"ciphertext"`
	Mac            ByteSlice `json:"mac"`
}

func (eo *EncryptOption) Decode(optStr string) *EncryptOption {
	if !Has0xPrefix(optStr) {
		optStr = "0x" + optStr
	}
	opts, err := hexutil.Decode(optStr)
	if err != nil {
		return nil
	}
	pubKey, err := btcec.ParsePubKey(opts[16:49], btcec.S256())
	if err != nil {
		return nil
	}
	return &EncryptOption{
		Iv:             opts[:16],                      // 16bytes
		EphemPublicKey: pubKey.SerializeUncompressed(), // 33bytes
		Mac:            opts[49:81],                    // 32bytes
		Ciphertext:     opts[81:],
	}
}

func (eo *EncryptOption) String() string {
	val, _ := json.Marshal(eo)
	return string(val)
}

func Has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func EncryptByPubKey(publicKey, message string) ([]byte, error) {
	if publicKey == "" || message == "" {
		return nil, fmt.Errorf("params invalidate for encryptWithPubKey")
	}

	signerKey, err := hexutil.Decode("0x" + string(publicKey))
	if err != nil {
		return nil, fmt.Errorf(`decode public key error:` + err.Error())
	}

	pubKey, err := btcec.ParsePubKey(signerKey, btcec.S256())
	if err != nil {
		return nil, fmt.Errorf(`parse public key error:` + err.Error())
	}

	encryptData, err := btcec.Encrypt(pubKey, []byte(message))
	if err != nil {
		return nil, fmt.Errorf(`encrypt data error:` + err.Error())
	}
	return encryptData, nil
}

// DecryptByKey front-end use https://github.com/pubkey/eth-crypto to encrypt value
func DecryptByKey(opt *EncryptOption, keyStr string) ([]byte, error) {
	if opt == nil {
		return nil, errors.New("EncryptOption error")
	}
	if !Has0xPrefix(keyStr) {
		keyStr = "0x" + keyStr
	}
	keyBytes, err := hexutil.Decode(keyStr)
	if err != nil {
		return nil, err
	}
	priKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)
	pubKey, _ := btcec.ParsePubKey(opt.EphemPublicKey, btcec.S256())
	ecdhKey := btcec.GenerateSharedSecret(priKey, pubKey)
	derivedKey := sha512.Sum512(ecdhKey)
	keyE := derivedKey[:32]
	keyM := derivedKey[32:]
	var dataToMac []byte
	dataToMac = append(dataToMac, opt.Iv...)
	dataToMac = append(dataToMac, append(opt.EphemPublicKey, opt.Ciphertext...)...)
	hm := hmac.New(sha256.New, keyM)
	hm.Write(dataToMac) // everything is hashed
	expectedMAC := hm.Sum(nil)
	if !hmac.Equal(opt.Mac, expectedMAC) {
		return nil, btcec.ErrInvalidMAC
	}
	// start decryption
	block, err := aes.NewCipher(keyE)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, opt.Iv)
	// same length as ciphertext
	plaintext := make([]byte, len(opt.Ciphertext))
	mode.CryptBlocks(plaintext, opt.Ciphertext)
	plain, err := removePKCSPadding(plaintext)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

// removePKCSPadding removes padding from data that was added with addPKCSPadding
func removePKCSPadding(src []byte) ([]byte, error) {
	length := len(src)
	padLength := int(src[length-1])
	if padLength > aes.BlockSize || length < aes.BlockSize {
		return nil, errors.New("errInvalidPadding")
	}
	return src[:length-padLength], nil
}
