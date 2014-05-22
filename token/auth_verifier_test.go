package token

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"
)

func TestAuthVerifier(t *testing.T) {
	v := NewVerifier("validToken")
	a := NewAuthVerifier("rainbow", v)
	now := time.Now()
	cases := []struct {
		token    string
		now      time.Time
		id       int
		hasError bool
	}{
		// when the token expired
		{
			encodeToken(v, []byte(fmt.Sprintf(`[123,"rainbow",%v]`, now.Unix()))),
			now,
			0,
			true,
		},
		// when the token is valid
		{
			encodeToken(v, []byte(fmt.Sprintf(`[123,"rainbow",%v]`, now.Add(1*time.Second).Unix()))),
			now,
			123,
			false,
		},
		// when name doesnt match
		{
			encodeToken(v, []byte(fmt.Sprintf(`[123,"unicorn",%v]`, now.Add(1*time.Second).Unix()))),
			now,
			0,
			true,
		},
	}

	for _, c := range cases {
		id, err := a.Verify(c.token, c.now)
		if (err != nil) != c.hasError || id != c.id {
			t.Fatal("Failed to verify", c, id, err)
		}
	}
}

func encodeToken(v *Verifier, data []byte) string {
	res := v.Generate(data)
	return base64.StdEncoding.EncodeToString([]byte(res))
}
