package mailbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	slogGorm "github.com/ajenpan/surf/core/utils/slog-gorm"
	msgMailbox "github.com/ajenpan/surf/msg/mailbox"
	gamedbMod "github.com/ajenpan/surf/server/mailbox/database/models"
)

var log = slog.Default().With("module", "mailbox")

type RecvMailMark = uint32

type MailID = uint64

const MailMarkRead RecvMailMark = 0b001   //1
const MailMarkRecv RecvMailMark = 0b010   //2
const MailMarkDelete RecvMailMark = 0b100 //4
const MailMarkMax RecvMailMark = 0b111

const MailMaxKeepCount = 20

const announcementFilePath = "./announcement.txt"

func createMysqlClient(dsn string) *gorm.DB {
	dbc, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableNestedTransaction: true, //关闭嵌套事务
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: slogGorm.New(slogGorm.WithLogger(slog.Default().With("module", "gorm"))),
	})
	if err != nil {
		log.Error("create mysql client error", "dsn", dsn, "error", err)
		panic(err)
	}
	return dbc
}

func NewHandler(c *Config) *Handler {
	ret := &Handler{
		conf: c,
	}

	ret.WLogDB = createMysqlClient(DefaultConf.WLogDBDSN)
	ret.WGameDB = createMysqlClient(DefaultConf.WGameDBDSN)
	ret.WPropsDB = createMysqlClient(DefaultConf.WPropsDBDSN)
	ret.Rds = redis.NewClient(&redis.Options{
		Addr:     c.RedisConn,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := ret.Init(); err != nil {
		log.Error("init mailbox error", "error", err)
		return nil
	}
	return ret
}

type User struct {
	UID uint32
}

type Handler struct {
	conf  *Config
	cache *MailCache

	WGameDB  *gorm.DB
	WPropsDB *gorm.DB
	WLogDB   *gorm.DB
	Rds      *redis.Client
}

func (h *Handler) Init() error {
	h.cache = &MailCache{
		infos: make(map[MailID]*MailDetail),
	}

	lists := []*gamedbMod.MailList{}

	err := h.WGameDB.Model(&gamedbMod.MailList{}).Order(gamedbMod.MailListColumns.Mailid + " DESC").Limit(MailMaxKeepCount * 10).Find(&lists).Error

	if err != nil {
		log.Error("init mailbox error", "error", err)
		return err
	}

	now := time.Now()
	details := []*MailDetail{}

	for _, record := range lists {
		detail, err := NewMailDetail(record)
		if err != nil {
			log.Error("init mailbox error", "error", err)
			continue
		}
		details = append(details, detail)
		// 如果邮件已经失效
		if now.After(detail.ExpireAt) && record.Status == 0 {
			err = h.WGameDB.Model(&gamedbMod.MailList{}).Where(&gamedbMod.MailList{Mailid: record.Mailid}).
				Update(gamedbMod.MailListColumns.Status, 2).Error
			if err != nil {
				log.Error("init mailbox error", "error", err)
			}
		}
	}

	err = h.cache.Add(details...)
	if err != nil {
		log.Error("init mailbox error", "error", err)
	}

	// if err := h.ann.Init(); err != nil {
	// 	return err
	// }

	return nil
}

func (h *Handler) RecvMail(ctx context.Context, in *msgMailbox.ReqRecvMail, out *msgMailbox.RespRecvMail) error {
	// get user info
	// user, err := GetUserFromCtx(ctx)
	// if err != nil {
	// 	return err
	// }

	//there's no new mail
	// if in.LatestMailid >= h.cache.LatestMailID() {
	// 	out.LatestCheckMailid = in.LatestMailid
	// 	return nil
	// }

	// get the latest mail id
	// RecvLatestMailID := database.UserRecvLatestMailID(user.UID)

	// checkPoint := in.LatestMailid
	// if in.LatestMailid < RecvLatestMailID {
	// 	checkPoint = RecvLatestMailID
	// }

	// // recv new mail
	// newMail := h.cache.RecvNewMail(checkPoint, user, MailMaxKeepCount)
	// if len(newMail) > 0 {

	// 	newRecord := []*gamedbMod.MailRecv{}

	// 	for _, v := range newMail {
	// 		if v.GetMailID() <= uint32(RecvLatestMailID) {
	// 			continue
	// 		}
	// 		newRecord = append(newRecord, &gamedbMod.MailRecv{
	// 			Mailid: uint(v.GetMailID()),
	// 			Mark:   0,
	// 			RecvAt: time.Now(),
	// 		})
	// 	}

	// 	if len(newRecord) > 0 {
	// 		err := h.WGameDB.Model(gamedbMod.MailRecv{}).CreateInBatches(newRecord, 10).Error
	// 		if err != nil {
	// 			log.Error(err)
	// 		}
	// 	}
	// }

	// lists := []*gamedbMod.MailRecv{}
	// err = h.WGameDB.Model(&gamedbMod.MailRecv{}).Order("mailid desc").Where("uid=? and mailid>? and status=0 and mark&4=0",
	// 	user.UID, in.LatestMailid).Limit(MailMaxKeepCount).Find(&lists).Error
	// if err != nil {
	// 	return err
	// }

	// for _, record := range lists {
	// 	mail := h.getMail(uint32(record.Mailid))
	// 	if mail == nil {
	// 		log.Warnf("mail not found, mailid:%d", record.Mailid)
	// 		continue
	// 	}

	// 	PBMail := mail.ClonePBMail()
	// 	PBMail.Mark = uint32(record.Mark)

	// 	mail.RWLock.RLock()
	// 	PBMail.RecvAt = mail.DBMail.CreateAt.Unix()
	// 	mail.RWLock.RUnlock()

	// 	out.Mails = append(out.Mails, PBMail)
	// }

	// out.LatestCheckMailid = h.cache.LatestMailID()
	return nil
}

func (h *Handler) getMail(mailid MailID) *MailDetail {
	mail := h.cache.GetMailDetail(mailid)
	if mail == nil {
		//get from database

		record := &gamedbMod.MailList{
			Mailid: uint(mailid),
		}
		var err error
		if err = h.WGameDB.Model(record).Take(record, record).Error; err != nil {
			log.Error("get mail error", "error", err)
			return nil
		}

		mail, err = NewMailDetail(record)
		if err != nil {
			log.Error("get mail error", "error", err)
			return nil
		}

		if mail != nil {
			h.cache.Add(mail)
		}
	}
	return mail
}

func (h *Handler) SendMail(ctx context.Context, in *msgMailbox.ReqSendMail, out *msgMailbox.RespSendMail) error {
	uid, err := GetAdminUIDFromCtx(ctx)
	if err != nil {
		return err
	}

	if in.RecvConds == nil || len(in.RecvConds.Items) == 0 {
		return fmt.Errorf("recv condition is required")
	}

	if len(in.Title) == 0 || len(in.Content) == 0 {
		return fmt.Errorf("title or content is required")
	}

	in.MailEffectAt = time.Now().Format(TimeLayout)

	if len(in.MailEffectAt) == 0 || len(in.MailExpireAt) == 0 {
		return fmt.Errorf("effect time is required")
	}

	expireAt, err := time.ParseInLocation(TimeLayout, in.MailExpireAt, time.Local)
	if err != nil {
		return fmt.Errorf("expire time format error,%v", err)
	}
	if time.Now().After(expireAt) {
		return fmt.Errorf("expire time must after effect time now:%v", time.Now())
	}

	detailRaw, err := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}.Marshal(in)

	if err != nil {
		return err
	}

	record := &gamedbMod.MailList{
		MailDetail: detailRaw,
		CreateAt:   time.Now(),
		CreateBy:   uid,
	}

	//first
	if err := h.WGameDB.Debug().Create(record).Error; err != nil {
		log.Error("send mail error", "error", err)
		return err
	}

	detail, err := NewMailDetail(record)
	if err != nil {
		log.Error("send mail error", "error", err)
		return err
	}

	out.Mailid = uint64(record.Mailid)
	if err := h.cache.Add(detail); err != nil {
		log.Error("send mail error", "error", err)
		return err
	}

	return nil
}

func (h *Handler) UserMarkMail(ctx context.Context, in *msgMailbox.ReqMarkUserMail, out *msgMailbox.RespMarkUserMail) error {

	// get user info
	user, err := GetUserFromCtx(ctx)
	if err != nil {
		return err
	}

	if out.Result == nil {
		out.Result = make(map[uint64]uint32)
	}

	for k, v := range in.Marks {
		out.Result[k] = 0

		if k == 0 || v == 0 || v > MailMarkMax {
			// log.Warn("recv mark key:%d, value:%d", k, v)
			log.Warn("recv mark error", "key", k, "value", v)
			continue
		}

		recvAble := false
		if v&MailMarkRecv == MailMarkRecv {
			//如果有领取标记, 则需要特别处理, 防止并发情况下用户多领
			res := h.WGameDB.Exec("update bfun_mail_recv set mark=mark|2 where uid = ? and mailid=? and mark&6=0", user.UID, k)
			if res.Error != nil {
				continue
			}
			recvAble = (res.RowsAffected == 1)
		}

		err := h.WGameDB.Exec("update bfun_mail_recv set mark=mark|? where uid = ? and mailid=?", v, user.UID, k).Error
		if err != nil {
			log.Error("user mark mail error", "error", err)
			continue
		}
		if recvAble {
			detail := h.cache.GetMailDetail(k)
			if detail == nil {
				log.Error("mail not found")
				continue
			}
		}

		out.Result[k] = v
	}

	return nil
}

func (h *Handler) MailList(ctx context.Context, in *msgMailbox.ReqMailDetailList, out *msgMailbox.RespMailDetailList) error {
	in.Page = CheckListPage(in.Page)
	total := int64(0)

	data := []*gamedbMod.MailList{}

	err := h.WGameDB.Debug().Model(gamedbMod.MailList{}).Count(&total).Limit(int(in.Page.PageSize)).
		Offset(int(in.Page.PageNum * in.Page.PageSize)).Order(gamedbMod.MailListColumns.Mailid + " DESC").Find(&data).Error
	if err != nil {
		return err
	}

	out.Total = uint32(total)

	mailids := []uint{}
	for _, v := range data {
		mailids = append(mailids, v.Mailid)
		temp := &msgMailbox.ReqSendMail{}
		err := protojson.Unmarshal(v.MailDetail, temp)
		if err != nil {
			log.Error("mail list error", "error", err)
			continue
		}

		out.Mails = append(out.Mails, &msgMailbox.RespMailDetailList_MailDetail{
			Mailid:     uint64(v.Mailid),
			Title:      temp.Title,
			Content:    temp.Content,
			Attachment: temp.Attachment,
			RecvConds:  temp.RecvConds,
			EffectAt:   temp.MailEffectAt,
			ExpireAt:   temp.MailExpireAt,
			CreateBy:   v.CreateBy,
			CreateAt:   v.CreateAt.Format(TimeLayout),
			Status:     int32(v.Status),
		})
	}
	type Count struct {
		Mailid uint32
		Count  uint32
	}

	ReadCount := []*Count{}
	RecvCount := []*Count{}

	if err := h.WGameDB.Debug().Raw("select mailid, ifnull(count(*),0) as Count from bfun_mail_recv where mailid in ? and mark&?=? group by mailid", mailids, MailMarkRead, MailMarkRead).Scan(&ReadCount).Error; err != nil {
		log.Error("mail list error", "error", err)
	}

	if err := h.WGameDB.Debug().Raw("select mailid, ifnull(count(*),0) as Count from bfun_mail_recv where mailid in ? and mark&?=? group by mailid", mailids, MailMarkRecv, MailMarkRecv).Scan(&RecvCount).Error; err != nil {
		log.Error("mail list error", "error", err)
	}

	convToMap := func(recv []*Count) map[uint32]uint32 {
		ret := make(map[uint32]uint32, len(recv))
		for _, v := range recv {
			ret[v.Mailid] = v.Count
		}
		return ret
	}

	ReadCountMap := convToMap(ReadCount)
	RecvCountMap := convToMap(RecvCount)

	for _, v := range out.Mails {
		if v.Statist == nil {
			v.Statist = &msgMailbox.RespMailDetailList_Statist{}
		}
		v.Statist.AttachRecv = RecvCountMap[uint32(v.Mailid)]
		v.Statist.MailRead = ReadCountMap[uint32(v.Mailid)]
	}
	return nil
}

func (h *Handler) OnSetMailStatus(ctx context.Context, in *msgMailbox.ReqSetMailStatus, out *msgMailbox.RespSetMailStatus) error {
	record := &gamedbMod.MailList{
		Mailid: uint(in.Mailid),
		Status: int(in.Status),
	}

	err := h.WGameDB.Model(record).Select(gamedbMod.MailListColumns.Status).Updates(record).Error
	if err != nil {
		return err
	}

	mail := h.cache.GetMailDetail(in.Mailid)
	if mail != nil {
		mail.RWLock.Lock()
		mail.DBMail.Status = int(in.Status)
		mail.RWLock.Unlock()
	}
	return nil
}
