package gorainbow

import (
	. "launchpad.net/gocheck"
)

type VerifierSuite struct{}

var _ = Suite(&VerifierSuite{})

// '["foo","bar"]' in base64
var message = "WyJmb28iLCJiYXIiXQ==--0898cb188ac378e6ba570f7fe2a946d2a51a4b13"

func (s *VerifierSuite) TestMessageVerifierVerify(c *C) {
	m := MessageVerifier{"foobar"}
	res, _ := m.Verify(message)
	c.Assert(string(res), Equals, "[\"foo\",\"bar\"]")
}

func (s *VerifierSuite) TestMessageVerifierGenerate(c *C) {
	m := MessageVerifier{"foobar"}
	res := m.Generate("[\"foo\",\"bar\"]")
	c.Assert(string(res), Equals, message)
}
