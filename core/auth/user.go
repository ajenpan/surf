package auth

type User interface {
	UserID() uint32
	UserRole() uint16
}
