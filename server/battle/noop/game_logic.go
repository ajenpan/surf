package noop

import (
	"time"

	"github.com/ajenpan/surf/server/battle"
)

func NewGameLogic() battle.Logic {
	return &NoopLogic{}
}

type NoopLogic struct {
}

func (gl *NoopLogic) OnInit(opts battle.LogicOpts) error {
	return nil
}
func (gl *NoopLogic) OnPlayerJoin(p []battle.Player) error {
	return nil
}

func (gl *NoopLogic) OnTick(time.Duration) {

}
func (gl *NoopLogic) OnReset() {

}
func (gl *NoopLogic) OnPlayerMessage(p battle.Player, syn uint32, msgid uint32, data []byte) {

}
func (gl *NoopLogic) OnCommand(topic string, data []byte) {

}
func (gl *NoopLogic) OnPlayerConnStatus(player battle.Player, enable bool) {

}
