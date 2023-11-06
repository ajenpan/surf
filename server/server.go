package server

import (
	"crypto/rsa"
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

var AuthPublicKey *rsa.PublicKey

func VerifyToken(pk *rsa.PublicKey, tokenRaw string) (uint64, string, string, error) {
	claims := make(jwt.MapClaims)
	token, err := jwt.ParseWithClaims(tokenRaw, claims, func(t *jwt.Token) (interface{}, error) {
		return pk, nil
	})
	if err != nil {
		return 0, "", "", err
	}
	if !token.Valid {
		return 0, "", "", fmt.Errorf("invalid token")
	}
	uidstr := claims["uid"]
	uname := claims["aud"]
	role := claims["rid"]
	uid, _ := strconv.ParseUint(uidstr.(string), 10, 64)

	return uid, uname.(string), role.(string), err
}
