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

	dummyToken := func(id int, name string, now time.Time) string {
		data := []byte(fmt.Sprintf(`[%d,"%s",%v]`, id, name, now.Unix()))
		res := v.Generate(data)
		return base64.StdEncoding.EncodeToString([]byte(res))
	}

	cases := []struct {
		token    string
		now      time.Time
		id       int
		hasError bool
	}{
		// when the token expired
		{dummyToken(123, "rainbow", now), now, 0, true},
		// when the token is valid
		{dummyToken(123, "rainbow", now.Add(1*time.Second)), now, 123, false},
		// when name doesnt match
		{dummyToken(123, "unicorn", now.Add(1*time.Second)), now, 0, true},
	}

	for _, c := range cases {
		id, err := a.Verify(c.token, c.now)
		if (err != nil) != c.hasError || id != c.id {
			t.Fatal("Failed to verify", c, id, err)
		}
	}
}
