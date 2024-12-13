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

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/utils/rsagen"

	xerr "github.com/ajenpan/surf/core/errors"
	msgUAuth "github.com/ajenpan/surf/msg/uauth"
	"github.com/ajenpan/surf/server/uauth/database/cache"
	"github.com/ajenpan/surf/server/uauth/database/models"
)

var RegUname = regexp.MustCompile(`^[a-zA-Z0-9_]{4,16}$`)

func New(privateRaw, publicRaw []byte) *UAuth {
	pk, err := rsagen.ParseRsaPrivateKeyFromPem(privateRaw)
	if err != nil {
		panic(err)
	}

	return &UAuth{
		cache:        cache.NewMemory(),
		privateKey:   pk,
		publicKeyRaw: publicRaw,
	}
}

type UAuth struct {
	privateKey   *rsa.PrivateKey
	publicKeyRaw []byte
	cache        cache.AuthCache

	DB   *gorm.DB
	surf *core.Surf
}

func (h *UAuth) OnInit(surf *core.Surf) error {
	h.surf = surf
	conf := DefaultConf

	h.DB = core.NewMysqlClient(conf.DBAddr)
	h.DB.AutoMigrate(models.Users{})

	core.HandleRequestFromHttp(surf, "/Login", h.OnReqLogin)
	core.HandleRequestFromHttp(surf, "/Verify", h.OnVerifyToken)
	core.HandleRequestFromHttp(surf, "/AnonymousLogin", h.OnAnonymousLogin)

	surf.HttpMux().HandleFunc("/publickey", func(w http.ResponseWriter, r *http.Request) {
		w.Write(h.publicKeyRaw)
	})
	return nil
}

func (h *UAuth) OnReady() error {
	return nil
}

func (h *UAuth) OnStop() error {
	return nil
}

func (h *UAuth) AuthWrapper(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorstr := r.Header.Get("Authorization")
		authorstr = strings.TrimPrefix(authorstr, "Bearer ")
		if authorstr == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, err := auth.VerifyToken(&h.privateKey.PublicKey, []byte(authorstr))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		f(w, r)
	}
}

func (h *UAuth) OnAnonymousLogin(ctx context.Context, in *msgUAuth.ReqAnonymousLogin, out *msgUAuth.RespAnonymousLogin) error {
	var err error

	if !RegUname.MatchString(in.Uname) {
		return xerr.New(int16(msgUAuth.ResponseFlag_PasswdWrong), "error")
	}

	if len(in.Passwd) < 6 {
		return xerr.New(int16(msgUAuth.ResponseFlag_PasswdWrong), "error")
	}

	user := &models.Users{
		Uname: in.Uname,
	}

	res := h.DB.Limit(1).Find(user, user)
	if err := res.Error; err != nil {
		return xerr.New(int16(msgUAuth.ResponseFlag_PasswdWrong), "error")
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
			return xerr.New(int16(msgUAuth.ResponseFlag_DataBaseErr), "create failed")
		}
		if res.RowsAffected == 0 {
			return xerr.New(int16(msgUAuth.ResponseFlag_DataBaseErr), "create failed")
		}
	} else {
		if user.Passwd != in.Passwd {
			return xerr.New(int16(msgUAuth.ResponseFlag_PasswdWrong), "error")
		}
		if user.Stat != 0 {
			return xerr.New(int16(msgUAuth.ResponseFlag_StatErr), "error")
		}
	}

	assess, err := auth.GenerateToken(h.privateKey, &auth.UserInfo{
		UId: uint32(user.UID),
	}, 0)

	if err != nil {
		return xerr.New(int16(msgUAuth.ResponseFlag_GenTokenErr), err.Error())
	}

	cacheInfo := &cache.AuthCacheInfo{
		User:         user,
		AssessToken:  assess,
		RefreshToken: uuid.NewString(),
	}

	if err = h.cache.StoreUser(context.Background(), cacheInfo, time.Hour); err != nil {
		return err
	}

	out.AssessToken = assess

	out.UserInfo = &msgUAuth.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
	return nil
}

func (h *UAuth) OnReqLogin(ctx context.Context, in *msgUAuth.ReqLogin, out *msgUAuth.RespLogin) error {
	if !RegUname.MatchString(in.Uname) {
		return xerr.New(int16(msgUAuth.ResponseFlag_PasswdWrong), "error")
	}

	if len(in.Passwd) < 6 {
		return fmt.Errorf("passwd is error")
	}

	user := &models.Users{
		Uname: in.Uname,
	}

	res := h.DB.Limit(1).Find(user, user)
	if err := res.Error; err != nil {
		return err
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("passwd is error")
	}

	if user.Passwd != in.Passwd {
		return fmt.Errorf("passwd is error")
	}

	if user.Stat != 0 {
		return fmt.Errorf("stat is error")
	}

	assess, err := auth.GenerateToken(h.privateKey, &auth.UserInfo{UId: uint32(user.UID)}, 0)
	if err != nil {
		return err
	}

	cacheInfo := &cache.AuthCacheInfo{
		User:         user,
		AssessToken:  assess,
		RefreshToken: uuid.NewString(),
	}

	if err = h.cache.StoreUser(context.Background(), cacheInfo, time.Hour); err != nil {
		slog.Error("store user err", "err", err)
	}

	out.AssessToken = assess
	out.UserInfo = &msgUAuth.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}

	return nil
}

func (h *UAuth) OnReqUserInfo(ctx context.Context, in *msgUAuth.ReqUserInfo, out *msgUAuth.RespUserInfo) error {
	user := &models.Users{
		UID: in.Uid,
	}
	uc := h.cache.FetchUser(context.Background(), in.Uid)
	if uc != nil {
		user = uc.User
	} else {
		res := h.DB.Limit(1).Find(user, user)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("user not found")
		}
		h.cache.StoreUser(context.Background(), &cache.AuthCacheInfo{User: user}, time.Hour)
	}
	out.Info = &msgUAuth.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
	return nil
}

func (h *UAuth) OnVerifyToken(ctx context.Context, in *msgUAuth.ReqVerifyToken, out *msgUAuth.RespVerifyToken) error {
	uinfo, err := auth.VerifyToken(&h.privateKey.PublicKey, []byte(in.AccessToken))
	if err != nil {
		out.Ok = false
		return nil
	}
	out.Ok = true

	c := h.cache.FetchUser(context.Background(), int64(uinfo.UId))
	if c == nil || c.User == nil {
		return nil
	}

	user := c.User
	out.UserInfo = &msgUAuth.UserInfo{
		Uid:     user.UID,
		Uname:   user.Uname,
		Stat:    int32(user.Stat),
		Created: user.CreateAt.Unix(),
	}
	return nil
}

func (h *UAuth) OnReqRegister(ctx context.Context, in *msgUAuth.ReqRegister, out *msgUAuth.RespRegister) error {
	var err error
	if !RegUname.MatchString(in.Uname) {
		return fmt.Errorf("invalid username")
	}

	if len(in.Passwd) < 6 {
		return fmt.Errorf("invalid password")
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
			return err
		}
		slog.Error("create user err", "err", res.Error)
		err = fmt.Errorf("server internal error")
		return err
	}

	if res.RowsAffected == 0 {
		err = fmt.Errorf("create user error")
		return err
	}

	out.Msg = "ok"
	return nil
}
