package secret

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Generate(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("length must be greater than 0")
	}

	secret := make([]byte, n)

	for i := range secret {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))

		if err != nil {
			return "", err
		}

		secret[i] = charset[num.Int64()]
	}

	return string(secret), nil
}
