package token

import "testing"

func TestVerifierVerify(t *testing.T) {
	v := NewVerifier("foobar")

	assertions := []struct {
		signed   string
		orig     string
		hasError bool
	}{
		{
			// '["foo","bar"]' in base64
			"WyJmb28iLCJiYXIiXQ==--0898cb188ac378e6ba570f7fe2a946d2a51a4b13",
			`["foo","bar"]`,
			false,
		},
		{
			"WyJmb28iLCJiYXIiXQ==--236i789834569723950781239048710293471024",
			"",
			true,
		},
		{
			"",
			"",
			true,
		},
	}

	for _, a := range assertions {
		res, err := v.Verify(a.signed)
		if (err != nil) != a.hasError || string(res) != a.orig {
			t.Fatal("Failed to verify", a, res, err)
		}
	}
}

func TestVerifierGenerate(t *testing.T) {
	v := NewVerifier("foobar")
	res := v.Generate([]byte(`["foo","bar"]`))
	expected := "WyJmb28iLCJiYXIiXQ==--0898cb188ac378e6ba570f7fe2a946d2a51a4b13"
	if res != expected {
		t.Fatal("Failed to generate", expected, res)
	}
}
