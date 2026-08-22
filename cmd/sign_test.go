package cmd

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

// signHT builds the HT-Signature header of a body, the way the API does:
// HMAC-SHA256 over "<t>." + the raw body, keyed with the whole secret.
func signHT(t *testing.T, secret, body string) string {
	t.Helper()
	timestamp := time.Now().Unix()
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%d.%s", timestamp, body)
	return fmt.Sprintf("HT-Signature: t=%d,v1=%s", timestamp, hex.EncodeToString(mac.Sum(nil)))
}
