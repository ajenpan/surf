package handler

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	log "github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/server/auth/common"
	msg "github.com/ajenpan/surf/server/auth/proto"
	"github.com/ajenpan/surf/server/auth/store/cache"
	"github.com/ajenpan/surf/server/auth/store/models"
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
	ct := calltable.ExtractParseGRpcMethod(msg.File_proto_auth_proto.Services(), ret)
	ret.ct = ct
	return ret
}

type Auth struct {
	AuthOptions
	ct *calltable.CallTable[string]
}

func (*Auth) Captcha(ctx context.Context, in *msg.CaptchaRequest) (*msg.CaptchaResponse, error) {
	return &msg.CaptchaResponse{}, nil
}

func (h *Auth) Login(ctx context.Context, in *msg.LoginRequest) (*msg.LoginResponse, error) {
	out := &msg.LoginResponse{}

	if !RegUname.MatchString(in.Uname) {
		out.Flag = msg.LoginResponse_UNAME_ERROR
		out.Msg = "please input right uname"
		return out, nil
	}

	if len(in.Passwd) < 6 {
		out.Flag = msg.LoginResponse_PASSWD_ERROR
		out.Msg = "passwd is required"
		return out, nil
	}

	user := &models.Users{
		Uname: in.Uname,
	}

	res := h.DB.Limit(1).Find(user, user)
	if err := res.Error; err != nil {
		out.Flag = msg.LoginResponse_FAIL
		out.Msg = "user not found"
		return nil, fmt.Errorf("server internal error")
	}

	if res.RowsAffected == 0 {
		out.Flag = msg.LoginResponse_UNAME_ERROR
		out.Msg = "user not exist"
		return out, nil
	}

	if user.Passwd != in.Passwd {
		out.Flag = msg.LoginResponse_PASSWD_ERROR
		return out, nil
	}

	if user.Stat != 0 {
		out.Flag = msg.LoginResponse_STAT_ERROR
		return out, nil
	}

	assess, err := common.GenerateToken(h.PK, &common.UserClaims{
		UID:   uint64(user.UID),
		UName: user.Uname,
		Role:  "user",
	})
	if err != nil {
		return nil, err
	}

	cacheInfo := &cache.AuthCacheInfo{
		User:         user,
		AssessToken:  assess,
		RefreshToken: uuid.NewString(),
	}

	if err = h.Cache.StoreUser(ctx, cacheInfo, time.Hour); err != nil {
		log.Error(err)
	}

	out.AssessToken = assess
	out.RefreshToken = cacheInfo.RefreshToken
	out.UserInfo = &msg.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
	return out, nil
}

func (h *Auth) Logout(ctx context.Context, in *msg.LogoutRequest) (*msg.LogoutResponse, error) {

	return nil, nil
}

func (*Auth) RefreshToken(ctx context.Context, in *msg.RefreshTokenRequest) (*msg.RefreshTokenResponse, error) {
	//TODO
	return nil, nil
}

func (h *Auth) UserInfo(ctx context.Context, in *msg.UserInfoRequest) (*msg.UserInfoResponse, error) {
	user := &models.Users{
		UID: in.Uid,
	}
	uc := h.Cache.FetchUser(ctx, in.Uid)
	if uc != nil {
		user = uc.User
	} else {
		res := h.DB.Limit(1).Find(user, user)
		if res.Error != nil {
			return nil, fmt.Errorf("server internal error: %v", res.Error)
		}
		if res.RowsAffected == 0 {
			return nil, fmt.Errorf("user no found")
		}

		h.Cache.StoreUser(ctx, &cache.AuthCacheInfo{User: user}, time.Hour)
	}

	out := &msg.UserInfoResponse{
		Info: &msg.UserInfo{
			Uid:     user.UID,
			Uname:   user.Uname,
			Stat:    int32(user.Stat),
			Created: user.CreateAt.Unix(),
		},
	}
	return out, nil
}

func (h *Auth) Register(ctx context.Context, in *msg.RegisterRequest) (*msg.RegisterResponse, error) {
	if !RegUname.MatchString(in.Uname) {
		return nil, nil
	}
	if len(in.Passwd) < 6 {
		return nil, fmt.Errorf("passwd is required")
	}

	user := &models.Users{
		Uname:    in.Uname,
		Passwd:   in.Passwd,
		Nickname: in.Nickname,
		Gender:   'X',
	}

	f := &models.Users{Uname: in.Uname}

	if res := h.DB.Find(f, f); res.RowsAffected > 0 {
		return nil, fmt.Errorf("user alread exist")
	}

	res := h.DB.Create(user)

	if res.Error != nil {
		log.Error(res.Error)
		return nil, fmt.Errorf("server internal error")
	}

	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("create user error")
	}

	return &msg.RegisterResponse{Msg: "ok"}, nil
}

func (h *Auth) PublicKeys(ctx context.Context, in *msg.PublicKeysRequest) (*msg.PublicKeysResponse, error) {
	return &msg.PublicKeysResponse{Keys: h.PublicKey}, nil
}

func (h *Auth) AnonymousLogin(ctx context.Context, in *msg.AnonymousLoginRequest) (*msg.LoginResponse, error) {

	return nil, nil
}

func (h *Auth) ModifyPasswd(ctx context.Context, in *msg.ModifyPasswdRequest) (*msg.ModifyPasswdResponse, error) {
	return nil, nil
}

func (h *Auth) AuthWrapper(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorstr := r.Header.Get("Authorization")
		authorstr = strings.TrimPrefix(authorstr, "Bearer ")
		if authorstr == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, err := common.VerifyToken(&h.PK.PublicKey, authorstr)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		f(w, r)
	}
}

// func (h *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	path := strings.TrimPrefix(r.URL.Path, "/auth/")
// 	if path == "" {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	log.Info(path)
// }
