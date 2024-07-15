package auth

type User interface {
	UserID() uint32
	UserName() string
	UserRole() uint32
}
