package models

import (
	"time"
)

// Users [...]
type Users struct {
	UID      int64     `gorm:"autoIncrement:true;primaryKey;column:uid;type:bigint;not null;comment:'用户唯一id'" json:"uid"`         // 用户唯一id
	Uname    string    `gorm:"unique;column:uname;type:varchar(64);not null;comment:'用户名'" json:"uname"`                          // 用户名
	Passwd   string    `gorm:"column:passwd;type:varchar(64);not null;default:'';comment:'密码'" json:"passwd"`                     // 密码
	Nickname string    `gorm:"column:nickname;type:varchar(64);not null;default:'';comment:'昵称'" json:"nickname"`                 // 昵称
	Avatar   string    `gorm:"column:avatar;type:varchar(1024);not null;default:'';comment:'头像'" json:"avatar"`                   // 头像
	Gender   int8      `gorm:"column:gender;type:tinyint;not null;default:0;comment:'性别'" json:"gender"`                          // 性别
	Phone    string    `gorm:"column:phone;type:varchar(32);not null;default:'';comment:'电话号码'" json:"phone"`                     // 电话号码
	Email    string    `gorm:"column:email;type:varchar(64);not null;default:'';comment:'电子邮箱'" json:"email"`                     // 电子邮箱
	Stat     int8      `gorm:"column:stat;type:tinyint;not null;default:0;comment:'状态码'" json:"stat"`                             // 状态码
	CreateAt time.Time `gorm:"column:create_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:'创建时间'" json:"create_at"` // 创建时间
	UpdateAt time.Time `gorm:"column:update_at;type:datetime;default:null;comment:'修改时间'" json:"update_at"`                       // 修改时间
}

// TableName get sql table name.获取数据库表名
func (m *Users) TableName() string {
	return "users"
}

// UsersColumns get sql column name.获取数据库列名
var UsersColumns = struct {
	UID      string
	Uname    string
	Passwd   string
	Nickname string
	Avatar   string
	Gender   string
	Phone    string
	Email    string
	Stat     string
	CreateAt string
	UpdateAt string
}{
	UID:      "uid",
	Uname:    "uname",
	Passwd:   "passwd",
	Nickname: "nickname",
	Avatar:   "avatar",
	Gender:   "gender",
	Phone:    "phone",
	Email:    "email",
	Stat:     "stat",
	CreateAt: "create_at",
	UpdateAt: "update_at",
}
