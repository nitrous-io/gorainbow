package verifier

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

type TokenVerifier struct {
	secret string
}

func (v *TokenVerifier) Verify(signed string) ([]byte, error) {
	if len(signed) > 4 {
		msgs := strings.Split(signed, "--")
		encoded := msgs[0]
		signature := msgs[1]
		expectedSig := signatureFor(v.secret, encoded)
		if signature == expectedSig {
			return base64.StdEncoding.DecodeString(encoded)
		} else {
			return []byte{}, fmt.Errorf("Verify: invalid msg: %v", signed)
		}
	}
	return []byte{}, fmt.Errorf("Verify: invalid msg: %v", signed)

}

func (v *TokenVerifier) Generate(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	signature := signatureFor(v.secret, encoded)
	return encoded + "--" + signature
}

func NewTokenVerifier(secret string) SignatureVerifier {
	return &TokenVerifier{secret}
}

func signatureFor(secret, msg string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}
