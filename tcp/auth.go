package tcp

import (
	"crypto/rsa"
	"fmt"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func RsaTokenAuth(pk *rsa.PublicKey) func(pkg *THVPacket) (*UserInfo, error) {
	return func(pkg *THVPacket) (*UserInfo, error) {
		claims := make(jwt.MapClaims)
		token, err := jwt.ParseWithClaims(string(pkg.Body), claims, func(t *jwt.Token) (interface{}, error) {
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
		if uidstr, has := claims["uid"]; has {
			uid, _ := strconv.ParseUint(uidstr.(string), 10, 64)
			ret.UId = uid
		}
		if role, has := claims["rid"]; has {
			ret.Role = role.(string)
		}
		return ret, nil
	}
}
