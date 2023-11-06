package websocket

type Packet struct {
	Name string
	Head []byte
	Body []byte
}
