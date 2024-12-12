// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameUserProp = "user_prop"

// UserProp mapped from table <user_prop>
type UserProp struct {
	UID      int32     `gorm:"column:uid;primaryKey;autoIncrement:true;comment:用户id" json:"uid"` // 用户id
	PropID   int32     `gorm:"column:prop_id;primaryKey;comment:道具" json:"prop_id"`              // 道具
	PropCnt  int64     `gorm:"column:prop_cnt;not null;comment:道具数量" json:"prop_cnt"`            // 道具数量
	UpdateAt time.Time `gorm:"column:update_at;not null;comment:最后更新时间" json:"update_at"`        // 最后更新时间
}

// TableName UserProp's table name
func (*UserProp) TableName() string {
	return TableNameUserProp
}
