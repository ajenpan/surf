package core

type Server interface {
	OnInit(*Surf) error
	OnReady()
	OnStop() error
}
