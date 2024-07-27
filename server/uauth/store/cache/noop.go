package cache

import (
	"context"

	"github.com/ajenpan/surf/server/uauth/store/models"
)

type Noop struct {
}

func (Noop) StoreUser(ctx context.Context, user *AuthCacheInfo) error { return nil }

func (Noop) FetchUser(ctx context.Context, uid int64) *AuthCacheInfo {
	return &AuthCacheInfo{
		User: &models.Users{UID: uid},
	}
}

func (Noop) FetchUserByName(ctx context.Context, uname string) *AuthCacheInfo {
	return &AuthCacheInfo{
		User: &models.Users{Uname: uname},
	}
}
