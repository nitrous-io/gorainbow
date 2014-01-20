package gorainbow

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

type (
	MessageVerifier struct {
		Secret string
	}

	AuthTokenVerifier struct {
		Name string
	}
)

func (m MessageVerifier) Verify(signed string) ([]byte, error) {
	if len(signed) > 4 {
		msgs := strings.Split(signed, "--")
		encoded := msgs[0]
		signature := msgs[1]
		expectedSig := signatureFor(m.Secret, encoded)
		if signature == expectedSig {
			return base64.StdEncoding.DecodeString(encoded)
		} else {
			return []byte{}, fmt.Errorf("Verify: invalid msg: %v", signed)
		}
	}
	return []byte{}, fmt.Errorf("Verify: invalid msg: %v", signed)
}

func (m MessageVerifier) Generate(data string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	signature := signatureFor(m.Secret, encoded)
	return encoded + "--" + signature
}

func signatureFor(secret, msg string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}
