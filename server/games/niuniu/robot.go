package niuniu

type Robot struct {

}

func (r *Robot) OnInit() error {
	return nil
}

func (r *Robot) String() string {
	return "niuniu_robot"
}

func (r *Robot) OnJoinDesk() {
	//TODO: ??
	//req := &pb.PlayerGameReadyRequest{Ready: 1}
	//r.SendMessage(req)
}


func (r *Robot) OnLeaveDesk() {

}

func (r *Robot) Ready(flag int) error { return nil }

func (r *Robot) Leave() {}

// func (r *Robot) OnMessage(message proto.Message) {
// defer func() {
// 	if r := recover(); r != nil {
// 		log.Error(r)
// 	}
// }()
// fullname := string(message.ProtoReflect().Descriptor().FullName())
// path := strings.Split(fullname, ".")
// route := ""
// if len(path) > 0 {
// 	route = path[len(path)-1]
// }

// log.Infof("Robot got msg uid:%v,route:%v", r.GetUserID(), route)

// desc := r.manager.callTable.Get(route)
// if desc == nil {
// 	log.Warnf("Robot can't find the method:%s", route)
// 	return
// }
// Robot, ok := r.manager.robots.Load(r.GetUserID())
// if !ok {
// 	log.Error("")
// 	return
// }
// args := []reflect.Value{reflect.ValueOf(Robot), reflect.ValueOf(message)}
// desc.method.Func.Call(args)
// }

//func (r *RobotCore) LoginGameResponse(resp *pb.LoginGameResponse)                 {}
//func (r *RobotCore) GameRoomListResponse(resp *pb.GameRoomListResponse)           {}
//func (r *RobotCore) JoinGameRoomResponse(resp *pb.JoinGameRoomResponse)           {}
//func (r *RobotCore) ListRoomPlayerResponse(resp *pb.ListRoomPlayerResponse)       {}
//func (r *RobotCore) CreatePrivateRoomResponse(resp *pb.CreatePrivateRoomResponse) {}
//func (r *RobotCore) JoinGameDeskResponse(resp *pb.JoinGameDeskResponse)           {}

// func (r *Robot) LeaveGameRoomResponse(resp *api.PlayerLeaveGameRoomResponse) {}
// func (r *Robot) LeaveGameDeskResponse(resp *api.PlayerLeaveGameDeskResponse) {}
// func (r *Robot) HelpJoinDeskResponse(resp *api.PlayerHelpJoinDeskResponse)   {}
// func (r *Robot) PlayerReconnectResponse(resp *api.PlayerReconnectResponse)   {}
// func (r *Robot) JoinGameDeskNotify(notify *api.PlayerJoinGameDeskNotify)     {}
// func (r *Robot) LeaveGameDeskNotify(notify *api.PlayerLeaveGameDeskNotify)   {}
// func (r *Robot) GameQuickSoundNotify(notify *api.GameQuickSoundNotify)       {}
// func (r *Robot) PlayerReconnectNotify(notify *api.PlayerReconnectNotify)     {}

// func (r *Robot) NNGameDeskInfoResponse(msg *nnpb.NNGameDeskInfoResponse)   {}
// func (r *Robot) NNPlayerBankerResponse(msg *nnpb.NNPlayerBankerResponse)   {}
// func (r *Robot) NNPlayerBetRateResponse(msg *nnpb.NNPlayerBetRateResponse) {}
// func (r *Robot) NNPlayerOutCardResponse(msg *nnpb.NNPlayerOutCardResponse) {}
// func (r *Robot) NNGameReconnectNotify(msg *nnpb.NNGameReconnectNotify)     {}

// func (r *Robot) NNGameStatusNotify(msg *nnpb.NNGameStatusNotify) {
// 	r.gameInfo.Status = msg.GameStatus

// 	switch msg.GameStatus {
// 	case nnpb.NNGameStatus_NN_GAME_UNKNOW:

// 	case nnpb.NNGameStatus_NN_GAME_IDLE:
// 		// NNGameStatus = 1  // 空闲,等待玩家准备
// 		//r.SendMessage(nnpb.NNgame)
// 	case nnpb.NNGameStatus_NN_GAME_COUNTDOWN:
// 		// NNGameStatus = 2  // 开始倒计时
// 	case nnpb.NNGameStatus_NN_GAME_BEGIN:
// 		// NNGameStatus = 3  // 开始
// 	case nnpb.NNGameStatus_NN_GAME_BANKER:
// 		// NNGameStatus = 4  // 抢庄
// 		req := &nnpb.NNPlayerBankerRequest{
// 			Rob: rand.Int31n(2) + 1,
// 		}
// 		r.SendMessage(req)
// 	case nnpb.NNGameStatus_NN_GAME_BANKER_NOTIFY:
// 		// NNGameStatus = 5  //通知庄

// 	case nnpb.NNGameStatus_NN_GAME_BET:

// 		// NNGameStatus = 6  // 下注
// 	case nnpb.NNGameStatus_NN_GAME_SEND:
// 		// NNGameStatus = 7  // 发牌
// 	case nnpb.NNGameStatus_NN_GAME_OUTCARD:
// 		req := &nnpb.NNPlayerOutCardRequest{
// 			OutCard: &nnpb.NNOutCardInfo{
// 				Type:  0,
// 				Cards: r.playerInfo.CardInfo,
// 			},
// 		}
// 		r.SendMessage(req)
// 		// NNGameStatus = 8  // 亮牌
// 	case nnpb.NNGameStatus_NN_GAME_TALLY:
// 		// NNGameStatus = 9  // 游戏结算
// 	case nnpb.NNGameStatus_NN_GAME_OVER:
// 		r.playerInfo.Reset()
// 		r.gameInfo.Reset()
// 		//	 	NNGameStatus = 10 // 游戏结束,清理桌子
// 	}
// }

// func (r *Robot) NNPlayerBankerNotify(msg *nnpb.NNPlayerBankerNotify) {

// }

// func (r *Robot) NNBankerSeatNotify(msg *nnpb.NNBankerSeatNotify) {
// 	if msg.SeatId == r.GetSeatID() {
// 		r.playerInfo.Banker = true
// 	} else {
// 		r.playerInfo.Banker = false
// 	}
// }

// func (r *Robot) NNPlayerBetRateNotify(msg *nnpb.NNPlayerBetRateNotify) {}
// func (r *Robot) NNPlayerCardInfoNotify(msg *nnpb.NNPlayerCardInfoNotify) {
// 	if r.GetSeatID() == msg.SeatId {
// 		r.playerInfo.CardInfo = msg.Cards
// 	}
// }
// func (r *Robot) NNPlayerOutCardNotify(msg *nnpb.NNPlayerOutCardNotify) {}
// func (r *Robot) NNPlayerTallyNotify(msg *nnpb.NNPlayerTallyNotify)     {}
