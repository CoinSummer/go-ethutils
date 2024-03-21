package go_ethutils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// 非对称加密算法，增强数据传输中的安全性
// js lib https://github.com/pubkey/eth-crypto

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

func Decode(optStr string) *EncryptOption {
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

func Has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

func compress(pubKey []byte) ([]byte, error) {
	if pubKey[0] != 0x04 || len(pubKey) != 65 {
		return nil, errors.New("invalid public key")
	}

	result := make([]byte, 33)
	if pubKey[64]%2 == 1 {
		result[0] = 0x03
	} else {
		result[0] = 0x02
	}

	copy(result[1:], pubKey[1:33])
	return result, nil
}

func (eo *EncryptOption) Stringify() (string, error) {
	compressedKey, err := compress(eo.EphemPublicKey)
	if err != nil {
		return "", err
	}

	data := make([]byte, 0)
	data = append(data, eo.Iv...)
	data = append(data, compressedKey...)
	data = append(data, eo.Mac...)
	data = append(data, eo.Ciphertext...)

	return hex.EncodeToString(data), nil
}

func EncryptByPubKey(publicKey, message string) (*EncryptOption, error) {
	if publicKey[:2] == "0x" {
		publicKey = publicKey[2:]
	}

	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}

	pubKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return nil, err
	}

	btcecPubKey := (*btcec.PublicKey)(pubKey)

	// Generate a random private/public key pair
	ephemeral, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}

	// Generate an ECDH shared secret
	sharedSecret := btcec.GenerateSharedSecret(ephemeral, btcecPubKey)
	// Hash the shared secret
	derivedKey := sha512.Sum512(sharedSecret)
	keyE := derivedKey[:32]
	keyM := derivedKey[32:]

	// Encrypt the message
	block, err := aes.NewCipher(keyE)
	if err != nil {
		return nil, err
	}

	// Generate a random initialization vector
	iv := make([]byte, aes.BlockSize)
	if _, err = rand.Read(iv); err != nil {
		return nil, err
	}

	// Use PKCS#7 padding
	messageBytes := []byte(message)
	messageBytes = pkcs7Padding(messageBytes, block.BlockSize())

	// Encrypt using AES in CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(messageBytes))
	mode.CryptBlocks(ciphertext, messageBytes)

	// Generate a MAC
	dataToMac := append(append(iv, ephemeral.PubKey().SerializeUncompressed()...), ciphertext...)
	hm := hmac.New(sha256.New, keyM)
	hm.Write(dataToMac)
	mac := hm.Sum(nil)

	return &EncryptOption{
		Iv:             iv,
		EphemPublicKey: ephemeral.PubKey().SerializeUncompressed(),
		Ciphertext:     ciphertext,
		Mac:            mac,
	}, nil
}

func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(ciphertext, padtext...)
}

func DecryptByKey(opt *EncryptOption, prvKeyStr string) ([]byte, error) {
	if opt == nil {
		return nil, errors.New("EncryptOption error")
	}
	if !Has0xPrefix(prvKeyStr) {
		prvKeyStr = "0x" + prvKeyStr
	}
	keyBytes, err := hexutil.Decode(prvKeyStr)
	if err != nil {
		return nil, fmt.Errorf("decode private key error: %v", err)
	}
	priKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)
	fmt.Printf("EphemPublicKey: %s\n", hex.EncodeToString(opt.EphemPublicKey))
	pubKey, err := btcec.ParsePubKey(opt.EphemPublicKey, btcec.S256())
	if err != nil {
		return nil, fmt.Errorf("parse public key error: %v", err)
	}

	pubKeyBytes := pubKey.SerializeCompressed() // 或者 pubKey.SerializeUncompressed()
	pubKeyStr := hex.EncodeToString(pubKeyBytes)
	fmt.Printf("pubKeyStr: %s\n", pubKeyStr)

	ecdhKey := btcec.GenerateSharedSecret(priKey, pubKey)
	fmt.Printf("ecdhKey: %x \n", ecdhKey)
	derivedKey := sha512.Sum512(ecdhKey)
	keyE := derivedKey[:32]
	keyM := derivedKey[32:]

	fmt.Printf("keyE: %x \n", keyE)
	fmt.Printf("keyM (Go): %x\n", keyM)

	var dataToMac []byte
	dataToMac = append(dataToMac, opt.Iv...)
	dataToMac = append(dataToMac, append(opt.EphemPublicKey, opt.Ciphertext...)...)
	fmt.Printf("dataToMac (Go): %x\n", dataToMac)
	hm := hmac.New(sha256.New, keyM)
	hm.Write(dataToMac) // everything is hashed
	expectedMAC := hm.Sum(nil)

	fmt.Printf("expectedMAC: %x \n", expectedMAC)
	fmt.Printf("opt.Mac: %x \n", opt.Mac)

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
		return nil, fmt.Errorf("invalid padding length: %d", padLength)
	}
	return src[:length-padLength], nil
}
