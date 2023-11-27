package auth

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func VerifyToken(pk *rsa.PublicKey, tokenRaw []byte) (*UserInfo, error) {
	claims := make(jwt.MapClaims)
	token, err := jwt.ParseWithClaims(string(tokenRaw), claims, func(t *jwt.Token) (interface{}, error) {
		return pk, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	ret := &UserInfo{}
	if uname, has := claims["aud"]; has {
		ret.UName = uname.(string)
	}
	if uid, has := claims["uid"]; has {
		ret.UId = uint32(uid.(float64))
	}
	if role, has := claims["rid"]; has {
		ret.URole = role.(string)
	}
	return ret, nil
}

func GenerateToken(pk *rsa.PrivateKey, uinfo *UserInfo, validity time.Duration) (string, error) {
	if validity == 0 {
		validity = 24 * time.Hour
	}
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	claims["iat"] = time.Now().Unix()
	claims["uid"] = float64(uinfo.UId)
	claims["aud"] = uinfo.UName
	claims["rid"] = uinfo.URole
	claims["iss"] = "hotwave"
	claims["sub"] = "auth"
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(pk)
}

func RsaTokenAuth(pk *rsa.PublicKey) func(data []byte) (*UserInfo, error) {
	return func(data []byte) (*UserInfo, error) {
		return VerifyToken(pk, data)
	}
}
