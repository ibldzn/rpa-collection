package fincloudapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func generateSignature(body []byte, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}
