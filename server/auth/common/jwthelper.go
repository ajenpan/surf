package common

import (
	"crypto/rsa"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// iss	Issuer	发行方
// sub	Subject	 主题
// aud	Audience	 受众
// exp	Expiration Time	过期时间
// nbf	Not Before	早于该定义的时间的JWT不能被接受处理
// iat	Issued At	JWT发行时的时间戳
// jti	JWT ID	JWT的唯一标识
// uid	用户ID
// rid	角色ID

type UserClaims struct {
	UID   uint64
	UName string
	Role  string
}

func GenerateToken(pk *rsa.PrivateKey, us *UserClaims) (string, error) {
	claims := make(jwt.MapClaims)
	claims["iss"] = "hotwave"
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	claims["aud"] = us.UName
	claims["uid"] = strconv.FormatInt(int64(us.UID), 10)
	claims["rid"] = us.Role
	// claims["sub"] = "auth"
	// claims["iat"] = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(pk)
}

func VerifyToken(pk *rsa.PublicKey, tokenRaw string) (*UserClaims, error) {
	claims := make(jwt.MapClaims)
	token, err := jwt.ParseWithClaims(tokenRaw, claims, func(t *jwt.Token) (interface{}, error) {
		return pk, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	ret := &UserClaims{}
	if v, has := claims["rid"]; has {
		vstr, ok := v.(string)
		if ok {
			ret.Role = vstr
		}
	}
	if v, has := claims["aud"]; has {
		vstr, ok := v.(string)
		if ok {
			ret.UName = vstr
		}
	}
	if v, has := claims["uid"]; has {
		vstr, ok := v.(string)
		if ok {
			ret.UID, _ = strconv.ParseUint(vstr, 10, 64)
		}
	}
	return ret, nil
}
