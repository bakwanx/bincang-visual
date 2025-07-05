package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
)

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	padding := int(data[length-1])
	if padding > length || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	return data[:(length - padding)], nil
}

func EncryptText(text string) (string, error) {
	privateKey := os.Getenv("PRIVATE_ENCRYPTOR_KEY")
	ivPrivateKey := os.Getenv("IV_PRIVATE_ENCRYPTOR_KEY")

	block, err := aes.NewCipher([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("error creating AES cipher: %v", err)
	}

	plainTextPadded := pkcs7Pad([]byte(text), block.BlockSize())
	ciphertext := make([]byte, len(plainTextPadded))

	mode := cipher.NewCBCEncrypter(block, []byte(ivPrivateKey))
	mode.CryptBlocks(ciphertext, plainTextPadded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptText(encrypted string) (string, error) {
	privateKey := os.Getenv("PRIVATE_ENCRYPTOR_KEY")
	ivPrivateKey := os.Getenv("IV_PRIVATE_ENCRYPTOR_KEY")

	block, err := aes.NewCipher([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("error creating AES cipher: %v", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("error decoding base64: %v", err)
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return "", fmt.Errorf("ciphertext not a multiple of block size")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, []byte(ivPrivateKey))
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("error unpadding: %v", err)
	}

	return string(plaintext), nil
}
