package niuniu

import (
	"time"

	"google.golang.org/protobuf/proto"

	log "github.com/ajenpan/surf/core/log"
)

func init() {
	// robotCallTable = gf.ParseProtoMessageWithSuffix("Response", api.File_openapi_gameframe_proto.Messages(), &DefaultRobotControl{})
	// gfNotifyCallTable := gf.ParseProtoMessageWithSuffix("Notify", api.File_openapi_gameframe_proto.Messages(), &DefaultRobotControl{})
	// nnCallTable := gf.ParseProtoMessageWithSuffix("Response", nnpb.File_niuniu_proto.Messages(), &DefaultRobotControl{})
	// nnNotifyCallTable := gf.ParseProtoMessageWithSuffix("Notify", nnpb.File_niuniu_proto.Messages(), &DefaultRobotControl{})
	// l := robotCallTable.Merge(nnCallTable, true)
	// if l != nnCallTable.Len() {
	// 	log.Error("init robot call table conflict")
	// }
	// l = robotCallTable.Merge(gfNotifyCallTable, true)
	// if l != nnCallTable.Len() {
	// 	log.Error("init robot call table conflict")
	// }
	// l = robotCallTable.Merge(nnNotifyCallTable, true)
	// if l != nnCallTable.Len() {
	// 	log.Error("init robot call table conflict")
	// }
	// log.Infof("init robot call tabel complete. table len:%v", robotCallTable.Len())
}

func NewDefaultRobotControl() *DefaultRobotControl {
	ret := &DefaultRobotControl{}
	return ret
}

type DefaultRobotControl struct {
	robots map[int64]*Robot

	action chan func()
}

func (r *DefaultRobotControl) work() {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		close(r.action)
		ticker.Stop()
	}()

	job := func(f func()) {
		defer func() {
			if r := recover(); r != nil {
				log.Error(r)
			}
		}()
		f()
	}
	for {
		select {
		case f, ok := <-r.action:
			if !ok {
				log.Errorf("room do action false")
				break
			}
			job(f)
		case _, ok := <-ticker.C:
			if !ok {
				log.Errorf("room do ticker false")
				break
			}
			job(r.OnTime)
		}
	}
}

func (m *DefaultRobotControl) OnTime() {

}

func (m *DefaultRobotControl) String() string {
	return "RobotControl"
}

func (m *DefaultRobotControl) Do(f func()) {
	m.action <- f
}

func (m *DefaultRobotControl) SendMessage() {

}

func (m *DefaultRobotControl) OnEvent(event proto.Message) {
	// m.action <- func() {
	// 	m.DealEvent(event)
	// }
}

func (m *DefaultRobotControl) OnRobotMessage() {

}

// func (m *DefaultRobotControl) DealEvent(event proto.Message) {
// 	switch event := event.(type) {
// 	case *pb.DeskEventPlayerJoin:
// 		m.OnDeskEventPlayerJoin(event)
// 	case *pb.DeskEventGameBegin:

// 	case *pb.DeskEventGameOver:

// 	case *pb.DeskEventPlayerLeave:
// 		m.OnDeskEventPlayerLeave(event)
// 	}
// }

func (m *DefaultRobotControl) CreateRobotAndJoin() {

}

func (m *DefaultRobotControl) RemoveAllRobots() {

}

func (m *DefaultRobotControl) RobotJoinDesk(r *Robot) {
	// req := &api.PlayerJoinGameDeskRequest{
	// 	DeskId: m.d.ID(),
	// }
	// m.SendMessage(r, req)
}

func (m *DefaultRobotControl) RobotReady(r *Robot, ready bool, data string) {

}

func (m *DefaultRobotControl) RobotLeaveDesk(r *Robot) {

}

// func (m *DefaultRobotControl) OnDeskEventPlayerLeave(event *pb.DeskEventPlayerLeave) {
// 	realPlayerCount := m.d.GetPlayerCount() - len(m.robots)
// 	//如果没有真玩家在座子上,移除所有机器人
// 	if realPlayerCount == 0 {
// 		m.RemoveAllRobots()
// 	}
// }

// func (m *DefaultRobotControl) OnDeskEventPlayerJoin(event *pb.DeskEventPlayerJoin) {

// 	//inDeskRobots := curDesk.GetRobots()

// 	robotCount := len(m.robots)

// 	realPlayerCount := m.d.GetPlayerCount() - robotCount

// 	if event.PlayerInfo.Robot {
// 		if realPlayerCount == 0 {
// 			//移除所有机器人
// 			m.RemoveAllRobots()
// 		}
// 		return
// 	}

// 	//如果桌子上已经有了1个机器人.不做任何处理
// 	if robotCount >= 1 {
// 		return
// 	}

// 	// 如果只有一个玩家,那么定时加入机器人
// 	if realPlayerCount == 1 {

// 		wait := rand.Int31n(5) + 1

// 		time.AfterFunc(time.Duration(wait)*time.Second, func() {
// 			m.action <- (func() {
// 				robotCount := len(m.robots)

// 				realPlayerCount := m.d.GetPlayerCount() - robotCount
// 				if realPlayerCount == 1 {
// 					m.CreateRobotAndJoin()
// 				}
// 			})
// 		})
// 	}
// }

// func (m *DefaultRobotControl) LeaveGameRoomResponse(r *Robot, resp *api.PlayerLeaveGameRoomResponse) {
// }
// func (m *DefaultRobotControl) LeaveGameDeskResponse(r *Robot, resp *api.PlayerLeaveGameDeskResponse) {
// }
// func (m *DefaultRobotControl) HelpJoinDeskResponse(r *Robot, resp *api.PlayerHelpJoinDeskResponse) {}
// func (m *DefaultRobotControl) PlayerReconnectResponse(r *Robot, resp *api.PlayerReconnectResponse) {}

// func (m *DefaultRobotControl) JoinGameDeskNotify(r *Robot, notify *api.PlayerJoinGameDeskNotify)   {}
// func (m *DefaultRobotControl) LeaveGameDeskNotify(r *Robot, notify *api.PlayerLeaveGameDeskNotify) {}
// func (m *DefaultRobotControl) GameQuickSoundNotify(r *Robot, notify *api.GameQuickSoundNotify)     {}
// func (m *DefaultRobotControl) PlayerReconnectNotify(r *Robot, notify *api.PlayerReconnectNotify)   {}

// func (m *DefaultRobotControl) NNGameDeskInfoResponse(r *Robot, msg *nnpb.NNGameDeskInfoResponse) {

// }

// func (m *DefaultRobotControl) NNPlayerBankerResponse(r *Robot, msg *nnpb.NNPlayerBankerResponse) {

// }
// func (m *DefaultRobotControl) NNPlayerBetRateResponse(r *Robot, msg *nnpb.NNPlayerBetRateResponse) {}
// func (m *DefaultRobotControl) NNPlayerOutCardResponse(r *Robot, msg *nnpb.NNPlayerOutCardResponse) {}

// func (m *DefaultRobotControl) NNGameReconnectNotify(r *Robot, msg *nnpb.NNGameReconnectNotify) {

// }
