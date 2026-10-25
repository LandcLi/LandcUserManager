// Package crypto 提供敏感字段加解密能力（如第三方登录 AccessToken）。
//
// 底层实现来自 landc-go tools/security（AES-GCM，密钥调用方传入）；
// 本包仅封装"密钥来源"：从环境变量 LUM_ENC_KEY 读取（未配置时退回开发密钥并告警，仅限开发环境）。
package crypto

import (
	"os"
	"sync"

	"github.com/LandcLi/landc-go/tools/security"
)

const (
	devKey       = "lum-dev-encryption-key-0123456789abcdef" // 32 bytes, 仅限开发
	encKeyEnvVar = "LUM_ENC_KEY"
)

var (
	keyOnce   sync.Once
	cipher    *security.Cipher
	cipherErr error
)

func getCipher() (*security.Cipher, error) {
	keyOnce.Do(func() {
		k := os.Getenv(encKeyEnvVar)
		if len(k) < 16 {
			// 未配置密钥：退回开发密钥（生产环境务必设置 LUM_ENC_KEY）
			println("WARNING: " + encKeyEnvVar + " not set, using dev encryption key (DO NOT use in production)")
			k = devKey
		}
		cipher, cipherErr = security.NewCipherFromString(k)
	})
	return cipher, cipherErr
}

// Encrypt AES-GCM 加密并返回带 enc: 前缀的 base64 字符串。
func Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	c, err := getCipher()
	if err != nil {
		return "", err
	}
	return c.Encrypt([]byte(plain))
}

// Decrypt 解密 enc: 前缀的密文；无前缀数据按明文原样返回（兼容历史数据）。
func Decrypt(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}
	c, err := getCipher()
	if err != nil {
		return "", err
	}
	b, err := c.Decrypt(encrypted)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
