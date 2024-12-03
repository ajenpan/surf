package uauth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	coreauth "github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/utils/calltable"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/errors"
	msgUAuth "github.com/ajenpan/surf/msg/uauth"
	"github.com/ajenpan/surf/server/uauth/database/cache"
	"github.com/ajenpan/surf/server/uauth/database/models"
)

var RegUname = regexp.MustCompile(`^[a-zA-Z0-9_]{4,16}$`)

type AuthOptions struct {
	PK        *rsa.PrivateKey
	PublicKey []byte
	DB        *gorm.DB
	Cache     cache.AuthCache
}

func NewAuth(opts AuthOptions) *Auth {
	ret := &Auth{
		AuthOptions: opts,
	}

	// 自动创建表
	opts.DB.AutoMigrate(models.Users{})
	return ret
}

type Auth struct {
	AuthOptions
}

func (h *Auth) ServerName() string {
	return "uauth"
}

func (h *Auth) ServerType() uint16 {
	return core.ServerType_UAuth
}

func (h *Auth) CTByName() *calltable.CallTable {
	ct := calltable.NewCallTable()
	ct.AddFunctionWithName("Login", h.OnReqLogin)
	ct.AddFunctionWithName("AnonymousLogin", h.AnonymousLogin)
	ct.AddFunctionWithName("Register", h.OnReqRegister)
	ct.AddFunctionWithName("UserInfo", h.OnReqUserInfo)
	return ct
}

func (h *Auth) AuthWrapper(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorstr := r.Header.Get("Authorization")
		authorstr = strings.TrimPrefix(authorstr, "Bearer ")
		if authorstr == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, err := coreauth.VerifyToken(&h.PK.PublicKey, []byte(authorstr))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		f(w, r)
	}
}

func (h *Auth) AnonymousLogin(ctx core.Context, in *msgUAuth.ReqAnonymousLogin) {
	out := &msgUAuth.RespAnonymousLogin{}
	var err error

	defer func() {
		ctx.Response(out, err)
	}()

	if !RegUname.MatchString(in.Uname) {
		err = errors.New(int32(msgUAuth.ResponseFlag_PasswdWrong), "error")
		return
	}

	if len(in.Passwd) < 6 {
		return
	}

	user := &models.Users{
		Uname: in.Uname,
	}

	res := h.DB.Limit(1).Find(user, user)
	if err := res.Error; err != nil {
		return
	}

	if res.RowsAffected == 0 {
		user = &models.Users{
			Uname:    in.Uname,
			Passwd:   in.Passwd,
			Nickname: "游客",
			Gender:   0,
		}
		res := h.DB.Create(user)
		if res.Error != nil {
			err = errors.New(int32(msgUAuth.ResponseFlag_DataBaseErr), "create failed")
		}
		if res.RowsAffected == 0 {
			err = errors.New(int32(msgUAuth.ResponseFlag_DataBaseErr), "create failed")
			return
		}
	} else {
		if user.Passwd != in.Passwd {
			err = errors.New(int32(msgUAuth.ResponseFlag_PasswdWrong), "error")
			return
		}
		if user.Stat != 0 {
			err = errors.New(int32(msgUAuth.ResponseFlag_StatErr), "error")
			return
		}
	}

	assess, err := coreauth.GenerateToken(h.PK, &coreauth.UserInfo{
		UId:   uint32(user.UID),
		UName: user.Uname,
	}, 0)

	if err != nil {
		err = errors.New(int32(msgUAuth.ResponseFlag_GenTokenErr), err.Error())
		return
	}

	cacheInfo := &cache.AuthCacheInfo{
		User:         user,
		AssessToken:  assess,
		RefreshToken: uuid.NewString(),
	}

	if err = h.Cache.StoreUser(context.Background(), cacheInfo, time.Hour); err != nil {
		return
	}

	out.AssessToken = assess

	out.UserInfo = &msgUAuth.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
}

func (h *Auth) OnReqLogin(ctx core.Context, in *msgUAuth.ReqLogin) {
	out := &msgUAuth.RespLogin{}
	// var err = &err.Error{}
	var err error

	defer func() {
		ctx.Response(out, err)
	}()

	if !RegUname.MatchString(in.Uname) {
		err = errors.New(int32(msgUAuth.ResponseFlag_PasswdWrong), "error")
		return
	}

	if len(in.Passwd) < 6 {
		return
	}

	user := &models.Users{
		Uname: in.Uname,
	}

	res := h.DB.Limit(1).Find(user, user)
	if err := res.Error; err != nil {
		return
	}

	if res.RowsAffected == 0 {
		return
	}

	if user.Passwd != in.Passwd {
		return
	}

	if user.Stat != 0 {
		return
	}

	assess, err := coreauth.GenerateToken(h.PK, &coreauth.UserInfo{
		UId:   uint32(user.UID),
		UName: user.Uname,
	}, 0)

	if err != nil {
		return
	}

	cacheInfo := &cache.AuthCacheInfo{
		User:         user,
		AssessToken:  assess,
		RefreshToken: uuid.NewString(),
	}

	if err = h.Cache.StoreUser(context.Background(), cacheInfo, time.Hour); err != nil {
		slog.Error("store user err", "err", err)
	}

	out.AssessToken = assess
	out.UserInfo = &msgUAuth.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
}

func (h *Auth) OnReqUserInfo(ctx core.Context, in *msgUAuth.ReqUserInfo) {
	user := &models.Users{
		UID: in.Uid,
	}
	uc := h.Cache.FetchUser(context.Background(), in.Uid)
	if uc != nil {
		user = uc.User
	} else {
		res := h.DB.Limit(1).Find(user, user)
		if res.Error != nil {
			ctx.Response(nil, res.Error)
			return
		}
		if res.RowsAffected == 0 {
			ctx.Response(nil, fmt.Errorf("user not found"))
			return
		}
		h.Cache.StoreUser(context.Background(), &cache.AuthCacheInfo{User: user}, time.Hour)
	}

	out := &msgUAuth.RespUserInfo{
		Info: &msgUAuth.UserInfo{
			Uid:     user.UID,
			Uname:   user.Uname,
			Stat:    int32(user.Stat),
			Created: user.CreateAt.Unix(),
		},
	}

	ctx.Response(out, nil)
}

func (h *Auth) OnReqRegister(ctx core.Context, in *msgUAuth.ReqRegister) {
	resp := &msgUAuth.RespRegister{}
	var err error

	defer func() {
		ctx.Response(resp, err)
	}()

	if !RegUname.MatchString(in.Uname) {
		err = fmt.Errorf("invalid username")
		return
	}

	if len(in.Passwd) < 6 {
		err = fmt.Errorf("invalid password")
		return
	}

	user := &models.Users{
		Uname:    in.Uname,
		Passwd:   in.Passwd,
		Nickname: in.Nickname,
		Gender:   'X',
	}

	res := h.DB.Create(user)

	if res.Error != nil {
		if res.Error == gorm.ErrDuplicatedKey {
			err = fmt.Errorf("username already exists")
			return
		}
		slog.Error("create user err", "err", res.Error)
		err = fmt.Errorf("server internal error")
		return
	}

	if res.RowsAffected == 0 {
		err = fmt.Errorf("create user error")
		return
	}

	resp.Msg = "ok"
}
