package core

type Server interface {
	OnInit(surf *Surf) error
	OnReady()
	OnStop() error
}
