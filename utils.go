package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"os"
	"runtime/debug"

	"github.com/polevpn/anyvalue"
	"github.com/polevpn/elog"
)

var ServerAesKey = []byte{0x75, 0xf3, 0xfe, 0x63, 0x18, 0x1f, 0x5c, 0x27, 0xab, 0x7c, 0xad, 0x4d, 0x7b, 0xf2, 0x59, 0xd0}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if length-unpadding <= 0 {
		return origData
	}
	return origData[:(length - unpadding)]
}

//AES加密
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//AES解密
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS7UnPadding(origData)
	return origData, nil
}

func GetConfig(configfile string) (*anyvalue.AnyValue, error) {

	f, err := os.Open(configfile)
	if err != nil {
		return nil, err
	}
	return anyvalue.NewFromJsonReader(f)
}

func PanicHandler() {
	if err := recover(); err != nil {
		elog.Error("Panic Exception:", err)
		elog.Error(string(debug.Stack()))
	}
}

func PanicHandlerExit() {
	if err := recover(); err != nil {
		elog.Error("Panic Exception:", err)
		elog.Error(string(debug.Stack()))
		elog.Error("************Program Exit************")
		elog.Flush()
		os.Exit(0)
	}
}
