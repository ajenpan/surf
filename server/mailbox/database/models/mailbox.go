package models

import (
	"time"

	"gorm.io/datatypes"
)

// BfunGiftCode [...]
type GiftCode struct {
	ID          int       `gorm:"primaryKey;column:id;type:int;not null" json:"id"`
	Code        string    `gorm:"unique;column:code;type:varchar(32);not null" json:"code"`
	GiftType    string    `gorm:"column:gift_type;type:varchar(32);not null" json:"gift_type"`
	GiftCount   int       `gorm:"column:gift_count;type:int;not null;default:0" json:"gift_count"`
	ValidCount  int       `gorm:"column:valid_count;type:int;not null;default:0" json:"valid_count"`
	RemainCount int       `gorm:"column:remain_count;type:int;not null;default:0" json:"remain_count"`
	ExpireAt    time.Time `gorm:"column:expire_at;type:datetime;not null" json:"expire_at"`
	CreateAt    time.Time `gorm:"column:create_at;type:datetime;not null;default:CURRENT_TIMESTAMP" json:"create_at"`
	Status      int8      `gorm:"column:status;type:tinyint;not null;default:0" json:"status"`
}

// TableName get sql table name.获取数据库表名
func (m *GiftCode) TableName() string {
	return "gift_code"
}

// BfunGiftCodeColumns get sql column name.获取数据库列名
var GiftCodeColumns = struct {
	ID          string
	Code        string
	GiftType    string
	GiftCount   string
	ValidCount  string
	RemainCount string
	ExpireAt    string
	CreateAt    string
	Status      string
}{
	ID:          "id",
	Code:        "code",
	GiftType:    "gift_type",
	GiftCount:   "gift_count",
	ValidCount:  "valid_count",
	RemainCount: "remain_count",
	ExpireAt:    "expire_at",
	CreateAt:    "create_at",
	Status:      "status",
}

// GiftExchangeColumns get sql column name.获取数据库列名
var GiftExchangeColumns = struct {
	ID       string
	Areaid   string
	Numid    string
	GiftCode string
	GiftStat string
	CreateAt string
}{
	ID:       "id",
	Areaid:   "areaid",
	Numid:    "numid",
	GiftCode: "gift_code",
	GiftStat: "gift_stat",
	CreateAt: "create_at",
}

// MailList [...]
type MailList struct {
	Mailid     uint           `gorm:"primaryKey;column:mailid;type:int unsigned;not null" json:"mailid"`                  // 邮件id
	MailDetail datatypes.JSON `gorm:"column:mail_detail;type:json;not null" json:"mail_detail"`                           // 内容
	CreateAt   time.Time      `gorm:"column:create_at;type:datetime;not null;default:CURRENT_TIMESTAMP" json:"create_at"` // 创建时间
	CreateBy   string         `gorm:"column:create_by;type:varchar(64);not null" json:"create_by"`                        // 创建人
	Status     int            `gorm:"column:status;type:int;not null;default:0" json:"status"`                            // 状态 0:正常, 1:失效
}

// TableName get sql table name.获取数据库表名
func (m *MailList) TableName() string {
	return "mail_list"
}

// MailListColumns get sql column name.获取数据库列名
var MailListColumns = struct {
	Mailid     string
	MailDetail string
	CreateAt   string
	CreateBy   string
	Status     string
}{
	Mailid:     "mailid",
	MailDetail: "mail_detail",
	CreateAt:   "create_at",
	CreateBy:   "create_by",
	Status:     "status",
}

// MailRecv [...]
type MailRecv struct {
	ID     uint64    `gorm:"primaryKey;column:id;type:bigint unsigned;not null" json:"id"`
	Areaid int       `gorm:"uniqueIndex:areaid_numid_mailid;column:areaid;type:int;not null" json:"areaid"`
	Numid  int       `gorm:"uniqueIndex:areaid_numid_mailid;column:numid;type:int;not null" json:"numid"`
	Mailid uint      `gorm:"uniqueIndex:areaid_numid_mailid;index:mailid;column:mailid;type:int unsigned;not null" json:"mailid"`
	Mark   uint      `gorm:"column:mark;type:int unsigned;not null;default:0" json:"mark"`
	RecvAt time.Time `gorm:"column:recv_at;type:datetime;not null" json:"recv_at"`
	Status int       `gorm:"column:status;type:int;not null;default:0" json:"status"`
}

// TableName get sql table name.获取数据库表名
func (m *MailRecv) TableName() string {
	return "mail_recv"
}

// BfunMailRecvColumns get sql column name.获取数据库列名
var BfunMailRecvColumns = struct {
	ID     string
	Areaid string
	Numid  string
	Mailid string
	Mark   string
	RecvAt string
	Status string
}{
	ID:     "id",
	Areaid: "areaid",
	Numid:  "numid",
	Mailid: "mailid",
	Mark:   "mark",
	RecvAt: "recv_at",
	Status: "status",
}
