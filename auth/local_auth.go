package auth

import (
	"crypto/rsa"
)

type LocalAuth struct {
	PK *rsa.PublicKey
}

func (a *LocalAuth) TokenAuth(token []byte) (*UserInfo, error) {
	return VerifyToken(a.PK, token)
}
