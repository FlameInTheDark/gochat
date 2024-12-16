package auth

import "github.com/golang-jwt/jwt/v5"

type Auth struct {
	Secret string
}

func New(secret string) *Auth {
	return &Auth{Secret: secret}
}

func (auth *Auth) signingFunc(token *jwt.Token) (interface{}, error) {
	return []byte(auth.Secret), nil
}

func (auth *Auth) Parse(token string) (*jwt.Token, error) {
	jwtoken, err := jwt.Parse(token, auth.signingFunc)
	if err != nil {
		return nil, err
	}
	return jwtoken, nil
}
