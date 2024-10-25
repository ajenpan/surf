package mailbox

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	pb "google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/log"
	msgMailbox "github.com/ajenpan/surf/msg/mailbox"
	gamedbMod "github.com/ajenpan/surf/server/mailbox/database/models"
)

type ReciChecker interface {
	Check(user *User) bool
}

type NumIDCheckerRange struct {
	MinUID uint32
	MaxUID uint32
}

func (c *NumIDCheckerRange) Check(user *User) bool {
	return user.UID >= c.MinUID && user.UID <= c.MaxUID
}

func NewNumIDListChecker(con string) (*UIDChecker, error) {
	ret := &UIDChecker{
		single: make(map[uint32]struct{}),
		ranges: []*NumIDCheckerRange{},
	}

	numids := strings.Split(con, ",")
	if len(numids) == 0 {
		return nil, fmt.Errorf("numid list must be split by ','")
	}

	for _, v := range numids {
		nn := strings.Split(v, "-")
		switch len(nn) {
		case 1:
			if n, err := strconv.Atoi(nn[0]); err != nil {
				return nil, err
			} else {
				ret.single[uint32(n)] = struct{}{}
			}
		case 2:
			var err error
			var min, max int
			if min, err = strconv.Atoi(nn[0]); err != nil {
				continue
			}
			if max, err = strconv.Atoi(nn[1]); err != nil {
				continue
			}
			ret.ranges = append(ret.ranges, &NumIDCheckerRange{
				MinUID: uint32(min),
				MaxUID: uint32(max),
			})

		}
	}
	return ret, nil
}

type UIDChecker struct {
	ranges []*NumIDCheckerRange
	single map[uint32]struct{}
}

func (c *UIDChecker) Check(user *User) bool {
	if _, has := c.single[user.UID]; has {
		return true
	}

	for _, v := range c.ranges {
		if v.Check(user) {
			return true
		}
	}
	return false
}

func PBMail2RecvMail(mailid MailID, info *msgMailbox.ReqSendMail) *msgMailbox.RespRecvMail_RecvMailInfo {
	ret := &msgMailbox.RespRecvMail_RecvMailInfo{
		Mailid:     mailid,
		Title:      info.Title,
		Content:    info.Content,
		Attachment: info.Attachment,
	}
	return ret
}

func MakeCheckerList(raw *msgMailbox.MailRecvCond) ([]ReciChecker, error) {
	ret := []ReciChecker{}
	for _, item := range raw.Items {
		switch item.Type {
		case msgMailbox.MailRecvCond_MailRecvCondItem_NumIDList:
			{
				checker, err := NewNumIDListChecker(item.Value)
				if err != nil {
					return nil, err
				}
				ret = append(ret, checker)
			}
		}
	}
	return ret, nil
}

func NewMailDetail(info *gamedbMod.MailList) (*MailDetail, error) {
	var err error

	ret := &MailDetail{
		DBMail:     *info,
		PBRecvMail: msgMailbox.RespRecvMail_RecvMailInfo{},
	}

	if err = protojson.Unmarshal(info.MailDetail, &ret.PBMail); err != nil {
		return nil, err
	}

	ret.PBRecvMail.Mailid = uint64(info.Mailid)
	ret.PBRecvMail.Title = ret.PBMail.Title
	ret.PBRecvMail.Content = ret.PBMail.Content
	ret.PBRecvMail.Attachment = ret.PBMail.Attachment

	ret.ExpireAt, err = time.ParseInLocation(TimeLayout, ret.PBMail.MailExpireAt, time.Local)
	if err != nil {
		return nil, err
	}

	ret.EffectAt, err = time.ParseInLocation(TimeLayout, ret.PBMail.MailEffectAt, time.Local)
	if err != nil {
		return nil, err
	}

	if ret.EffectAt.After(ret.ExpireAt) {
		return nil, fmt.Errorf("expire time must after effect time, mailid:%d must be clean", info.Mailid)
	}

	checks, err := MakeCheckerList(ret.PBMail.RecvConds)
	if err != nil {
		return nil, err
	}

	ret.Checkers = checks
	return ret, nil
}

type MailDetail struct {
	DBMail     gamedbMod.MailList
	PBMail     msgMailbox.ReqSendMail
	PBRecvMail msgMailbox.RespRecvMail_RecvMailInfo
	Checkers   []ReciChecker
	ExpireAt   time.Time
	EffectAt   time.Time
	RWLock     sync.RWMutex
}

func (m *MailDetail) ClonePBMail() *msgMailbox.RespRecvMail_RecvMailInfo {
	m.RWLock.RLock()
	defer m.RWLock.RUnlock()
	return pb.Clone(&m.PBRecvMail).(*msgMailbox.RespRecvMail_RecvMailInfo)
}

func (m *MailDetail) GetMailID() uint32 {
	m.RWLock.RLock()
	defer m.RWLock.RUnlock()
	return uint32(m.DBMail.Mailid)
}

func (m *MailDetail) Valid() bool {
	m.RWLock.RLock()
	defer m.RWLock.RUnlock()
	if m.DBMail.Status != 0 {
		return false
	}
	return m.ExpireAt.After(time.Now())
}

func (m *MailDetail) CheckRecv(user *User) bool {
	m.RWLock.RLock()
	defer m.RWLock.RUnlock()

	// 只要满足一条即可
	for _, checker := range m.Checkers {
		if checker.Check(user) {
			return true
		}
	}
	return false
}

type MailCache struct {
	rwlock    sync.RWMutex
	infos     map[MailID]*MailDetail
	infosKeys []MailID
}

func (c *MailCache) GetMailDetail(mid MailID) *MailDetail {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	if v, has := c.infos[mid]; has {
		return v
	}
	return nil
}

func (c *MailCache) Add(details ...*MailDetail) error {
	if len(details) == 0 {
		return nil
	}

	c.rwlock.Lock()
	defer c.rwlock.Unlock()

	for _, detail := range details {
		mailid := MailID(detail.DBMail.Mailid)
		if _, has := c.infos[mailid]; has {
			err := fmt.Errorf("mail already exist")
			log.Error(err)
			continue
		}

		if _, has := c.infos[mailid]; has {
			log.Warn("mail already exist,mailid:", mailid)
			continue
		}

		c.infos[mailid] = detail
		c.infosKeys = append(c.infosKeys, mailid)
	}

	sort.Slice(c.infosKeys, func(i, j int) bool {
		return c.infosKeys[i] < c.infosKeys[j]
	})

	return nil
}

func (c *MailCache) RecvNewMail(latestMailid uint64, user *User, limit uint32) []*MailDetail {
	if latestMailid >= c.LatestMailID() {
		return nil
	}

	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	ret := []*MailDetail{}

	for i := len(c.infosKeys) - 1; i >= 0; i-- {
		mailid := c.infosKeys[i]

		if mailid < latestMailid {
			// there is no new mail
			break
		}

		mail, has := c.infos[mailid]
		if !has {
			log.Error("mail not found by id:", mailid)
			continue
		}

		if !mail.Valid() {
			continue
		}

		if !mail.CheckRecv(user) {
			continue
		}

		ret = append(ret, mail)
		limit = limit - 1
		if limit == 0 {
			break
		}
	}

	return ret
}

func (c *MailCache) LatestMailID() MailID {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	if len(c.infosKeys) == 0 {
		return 0
	}
	return c.infosKeys[len(c.infosKeys)-1]
}
