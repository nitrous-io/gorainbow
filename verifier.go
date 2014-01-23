package gorainbow

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type (
	MessageVerifier struct {
		Secret string
	}

	AuthTokenVerifier struct {
		Name string
		m    MessageVerifier
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

func (m MessageVerifier) Generate(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	signature := signatureFor(m.Secret, encoded)
	return encoded + "--" + signature
}

// return userId embedded in
func (a AuthTokenVerifier) Verify(token []byte, now time.Time) (int, error) {
	return 0, nil
}

func signatureFor(secret, msg string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}
