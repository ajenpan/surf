package cache

import (
	"context"
	"sync"
	"time"
)

func NewMemory() AuthCache {
	return &Memory{
		cache:    make(map[int64]*AuthCacheInfo),
		name2id:  make(map[string]int64),
		token2id: make(map[string]int64),
	}
}

type Memory struct {
	rwLock   sync.RWMutex
	cache    map[int64]*AuthCacheInfo
	name2id  map[string]int64
	token2id map[string]int64
}

func (m *Memory) StoreUser(ctx context.Context, user *AuthCacheInfo, exprieAt time.Duration) error {
	m.DeleteUser(ctx, user.User.UID)

	m.rwLock.Lock()
	defer m.rwLock.Unlock()

	m.cache[user.User.UID] = user
	m.name2id[user.User.Uname] = user.User.UID
	m.token2id[user.AssessToken] = user.User.UID
	return nil
}

func (m *Memory) DeleteUser(ctx context.Context, uid int64) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()

	if u, has := m.cache[uid]; has {
		delete(m.cache, uid)
		delete(m.name2id, u.User.Uname)
		delete(m.token2id, u.AssessToken)
	}
}

func (m *Memory) FetchUser(ctx context.Context, uid int64) *AuthCacheInfo {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	return m.cache[uid]
}

func (m *Memory) FetchUserByName(ctx context.Context, uname string) *AuthCacheInfo {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	uid := m.name2id[uname]
	return m.cache[uid]
}

func (m *Memory) FetchUserByToken(ctx context.Context, AccessToken string) *AuthCacheInfo {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	uid, has := m.token2id[AccessToken]
	if has {
		return m.cache[uid]
	}
	return nil
}
