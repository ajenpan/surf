package core

type Server interface {
	OnInit(surf *Surf) error
	OnReady() error
	OnStop() error
}
