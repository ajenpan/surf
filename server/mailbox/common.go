package mailbox

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"

	pb "github.com/ajenpan/surf/msg/openproto/mailbox"
)

type (
	CtxClaimsKey   struct{}
	CtxXRealIpKey  struct{}
	CtxAdminUIDKey struct{}
	CtxUserKey     struct{}
	CtxCallerKey   struct{}
)

var (
	CtxAdminUID   = CtxAdminUIDKey{}
	CtxXRealIp    = CtxXRealIpKey{}
	CtxClaims     = CtxClaimsKey{}
	CtxUser       = CtxUserKey{}
	CtxCallerRole = CtxCallerKey{}
)

const TimeLayout = "2006-01-02 15:04:05"

var AdminAuthsigned = []byte("49370808e359957960ae9058f3860c9b")

func VerifyAdminTokenWithClaims(tokenRaw string) (map[string]interface{}, error) {
	claims := make(jwt.MapClaims)

	_, err := jwt.ParseWithClaims(tokenRaw, claims, func(t *jwt.Token) (interface{}, error) {
		return AdminAuthsigned, nil
	})

	return claims, err
}

func GetAdminUIDFromCtx(ctx context.Context) (string, error) {
	raw := ctx.Value(CtxAdminUID)
	if raw == nil {
		return "", fmt.Errorf("ctx without uid")
	}
	ret, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("GetUIDFromCtx failed")
	}
	return ret, nil
}

func GetCallerRoleFromCtx(ctx context.Context) (string, error) {
	raw := ctx.Value(CtxCallerRole)
	if raw == nil {
		return "", fmt.Errorf("ctx without caller")
	}
	ret, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("GetCallerFromCtx failed")
	}
	return ret, nil
}

func GetUserFromCtx(ctx context.Context) (*User, error) {
	raw := ctx.Value(CtxUser)
	if raw == nil {
		return nil, fmt.Errorf("ctx without uid")
	}
	ret, ok := raw.(*User)
	if !ok {
		return nil, fmt.Errorf("GetUIDFromCtx failed")
	}
	return ret, nil
}

func GetClaimsFromCtx(ctx context.Context) (jwt.MapClaims, error) {
	raw := ctx.Value(CtxClaims)
	if raw == nil {
		return nil, fmt.Errorf("ctx without claims")
	}

	ret, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("GetClaimsFromCtx failed")
	}
	return ret, nil
}

func GetXRealIpFromCtx(ctx context.Context) (string, error) {
	raw := ctx.Value(CtxXRealIp)
	if raw == nil {
		return "", fmt.Errorf("ctx without xrealip")
	}
	ret, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("GetXRealIpFromCtx failed")
	}
	return ret, nil
}

func CheckListPage(page *pb.ListPage) *pb.ListPage {
	if page == nil {
		page = &pb.ListPage{
			PageSize: 20,
			PageNum:  0,
		}
	} else {
		if page.PageSize > 100 || page.PageSize <= 0 {
			page.PageSize = 20
		}
		page.PageNum -= 1
		if page.PageNum < 0 {
			page.PageNum = 0
		}
	}
	return page
}

func RandStr(length int) string {
	rd := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	var letters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rd.Intn(len(letters))]
	}
	return string(b)
}
