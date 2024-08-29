package auth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	coreauth "github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/utils/calltable"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/errors"
	log "github.com/ajenpan/surf/core/log"
	pb "github.com/ajenpan/surf/msg/openproto/uauth"
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

func (h *Auth) AnonymousLogin(ctx core.Context, in *pb.AnonymousLoginRequest) {
	out := &pb.AnonymousLoginResponse{}
	var err error

	defer func() {
		ctx.Response(out, err)
	}()

	if !RegUname.MatchString(in.Uname) {
		err = errors.New(int32(pb.ResponseFlag_PasswdWrong), "error")
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
			err = errors.New(int32(pb.ResponseFlag_DataBaseErr), "create failed")
		}
		if res.RowsAffected == 0 {
			err = errors.New(int32(pb.ResponseFlag_DataBaseErr), "create failed")
			return
		}
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
		err = errors.New(int32(pb.ResponseFlag_GenTokenErr), err.Error())
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

	out.UserInfo = &pb.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
}

func (h *Auth) CTByName() *calltable.CallTable[string] {

	NewMethod := func(f interface{}) *calltable.Method {
		refv := reflect.ValueOf(f)
		if refv.Kind() != reflect.Func {
			return nil
		}
		ret := &calltable.Method{
			Func:        refv,
			RequestType: refv.Type().In(1).Elem(),
		}
		return ret
	}

	ct := calltable.NewCallTable[string]()
	ct.Add("Login", NewMethod(h.Login))
	ct.Add("AnonymousLogin", NewMethod(h.AnonymousLogin))
	return ct
}

func (h *Auth) Login(ctx core.Context, in *pb.LoginRequest) {
	out := &pb.LoginResponse{}
	// var err = &err.Error{}
	var err error

	defer func() {
		ctx.Response(out, err)
	}()

	if !RegUname.MatchString(in.Uname) {
		err = errors.New(int32(pb.ResponseFlag_PasswdWrong), "error")
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
		log.Error(err)
	}

	out.AssessToken = assess
	out.UserInfo = &pb.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
}

func (h *Auth) UserInfo(ctx core.Context, in *pb.UserInfoRequest) {
	user := &models.Users{
		UID: in.Uid,
	}
	uc := h.Cache.FetchUser(context.Background(), in.Uid)
	if uc != nil {
		user = uc.User
	} else {
		res := h.DB.Limit(1).Find(user, user)
		if res.Error != nil {
			return
		}
		if res.RowsAffected == 0 {
			return
		}
		h.Cache.StoreUser(context.Background(), &cache.AuthCacheInfo{User: user}, time.Hour)
	}

	out := &pb.UserInfoResponse{
		Info: &pb.UserInfo{
			Uid:     user.UID,
			Uname:   user.Uname,
			Stat:    int32(user.Stat),
			Created: user.CreateAt.Unix(),
		},
	}

	ctx.Response(out, nil)
}

func (h *Auth) Register(ctx core.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
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

	res := h.DB.Create(user)

	if res.Error != nil {
		log.Error(res.Error)
		return nil, fmt.Errorf("server internal error")
	}

	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("create user error")
	}

	return &pb.RegisterResponse{Msg: "ok"}, nil
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
