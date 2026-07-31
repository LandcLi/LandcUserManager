// Package crypto 提供敏感字段加解密能力（如第三方登录 AccessToken）
// 密钥从环境变量 LUM_ENC_KEY 读取，未配置时退回开发密钥并打印告警（仅限开发环境）
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"sync"
)

const (
	devKey       = "lum-dev-encryption-key-0123456789abcdef" // 32 bytes, 仅限开发
	encPrefix    = "enc:"
	encKeyEnvVar = "LUM_ENC_KEY"
)

var (
	keyOnce sync.Once
	key     []byte
)

func getKey() []byte {
	keyOnce.Do(func() {
		if k := os.Getenv(encKeyEnvVar); len(k) >= 32 {
			key = []byte(k[:32])
			return
		}
		// 未配置密钥：退回开发密钥（生产环境务必设置 LUM_ENC_KEY）
		println("WARNING: " + encKeyEnvVar + " not set, using dev encryption key (DO NOT use in production)")
		key = []byte(devKey)
	})
	return key
}

// Encrypt AES-GCM 加密并返回 base64 字符串，带 enc: 前缀标识
func Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}

	block, err := aes.NewCipher(getKey())
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 enc: 前缀的密文；非 enc: 前缀按明文原样返回（兼容历史数据）
func Decrypt(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}
	if len(encrypted) < len(encPrefix) || encrypted[:len(encPrefix)] != encPrefix {
		return encrypted, nil
	}

	raw, err := base64.StdEncoding.DecodeString(encrypted[len(encPrefix):])
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(getKey())
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	plain, err := gcm.Open(nil, raw[:nonceSize], raw[nonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
