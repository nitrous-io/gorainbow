package verifier

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type AuthVerifier struct {
	name     string
	verifier SignatureVerifier
}

func (a *AuthVerifier) Verify(b64Token string, now time.Time) (int, error) {
	ub64, err := base64.URLEncoding.DecodeString(b64Token)
	if err != nil {
		return 0, fmt.Errorf("invalid base64 token %v, err: $v", b64Token, err)
	}
	verifyResp, err := a.verifier.Verify(string(ub64))
	if err != nil {
		return 0, fmt.Errorf("failed to verify %v, err: %v", string(ub64), err)
	}
	var decoded []interface{}
	if err = json.Unmarshal(verifyResp, &decoded); err != nil {
		return 0, fmt.Errorf("failed to decode %v, err: %v", string(ub64), err)
	}
	id := int(decoded[0].(float64))
	name := decoded[1].(string)
	expiresAt := int64(decoded[2].(float64))
	if name != a.name || expiresAt <= now.Unix() {
		return 0, fmt.Errorf("invalid name %v or expiresAt %v", name, expiresAt)
	}
	return id, nil
}

func NewAuthVerifier(name string, v SignatureVerifier) UserVerifier {
	return &AuthVerifier{name, v}
}
