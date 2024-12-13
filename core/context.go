package core

import (
	"context"

	"github.com/ajenpan/surf/core/auth"
)

type ctxKey int

const (
	keyUser ctxKey = iota
	keyConnId
)

func CtxWithUser(ctx context.Context, u auth.User) context.Context {
	return context.WithValue(ctx, keyUser, u)
}

func CtxToUser(ctx context.Context) (auth.User, bool) {
	u, ok := ctx.Value(keyUser).(auth.User)
	return u, ok
}

func CtxToUId(ctx context.Context) uint32 {
	u, ok := ctx.Value(keyUser).(auth.User)
	if !(ok) {
		return 0
	}
	return u.UserID()
}

func CtxToURole(ctx context.Context) uint16 {
	u, ok := ctx.Value(keyUser).(auth.User)
	if !(ok) {
		return 0
	}
	return u.UserRole()
}

func CtxWithConnId(ctx context.Context, connid string) context.Context {
	return context.WithValue(ctx, keyConnId, connid)
}

func CtxToConnId(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(keyUser).(string)
	return u, ok
}
