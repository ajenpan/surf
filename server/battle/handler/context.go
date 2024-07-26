package handler

type userInfoKey struct{}
type tcpSocketKey struct{}

var UserInfoKey = &userInfoKey{}
var TcpSocketKey = &tcpSocketKey{}

// func GetUserInfo(ctx context.Context) *network.UserInfo {
// 	return ctx.Value(UserInfoKey).(*network.UserInfo)
// }

// func WithUserInfo(ctx context.Context, uinfo *tcp.UserInfo) context.Context {
// 	return context.WithValue(ctx, UserInfoKey, uinfo)
// }

// func GetTcpSocket(ctx context.Context) network.Conn {
// 	return ctx.Value(TcpSocketKey).(*tcp.Socket)
// }

// func WithTcpSocket(ctx context.Context, s *tcp.Socket) context.Context {
// 	return context.WithValue(ctx, TcpSocketKey, s)
// }
