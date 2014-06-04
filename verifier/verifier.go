package verifier

import "time"

// verify that request comes from the user and returns the user id
type UserVerifier interface {
	Verify(b64Token string, now time.Time) (int, error)
}

type SignatureVerifier interface {
	Verify(signed string) ([]byte, error)
	Generate(data []byte) string
}
