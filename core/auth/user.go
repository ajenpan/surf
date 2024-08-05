package auth

type UserRoleEnum = uint8

const (
	UserRole_User   UserRoleEnum = 0
	UserRole_Server UserRoleEnum = 1
)

type User interface {
	UserID() uint32
	UserName() string
	UserRole() uint32
}
