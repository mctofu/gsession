package gsession

import "github.com/gorilla/securecookie"

// CookieCodec can provide further processing of a session id when persisting
// it to a cookie.
type CookieCodec interface {
	Encode(name, value string) (string, error)
	Decode(name, value string) (string, error)
}

type SecureCookieCodec struct {
	Codecs []securecookie.Codec
}

func (c *SecureCookieCodec) Encode(name string, value string) (string, error) {
	return securecookie.EncodeMulti(name, value, c.Codecs...)
}

func (c *SecureCookieCodec) Decode(name string, value string) (string, error) {
	var decoded string
	err := securecookie.DecodeMulti(name, value, &decoded, c.Codecs...)
	return decoded, err
}
