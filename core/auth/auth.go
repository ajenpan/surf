package auth

type UserInfo struct {
	UId   uint32 `json:"uid"`
	UName string `json:"uname"`
	URole uint8  `json:"urid"`
}

func (u *UserInfo) UserID() uint32 {
	return u.UId
}
func (u *UserInfo) UserRole() uint8 {
	return u.URole
}
func (u *UserInfo) UserName() string {
	return u.UName
}
