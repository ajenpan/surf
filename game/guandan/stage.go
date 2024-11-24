package guandan

import "time"

type StageInfo struct {
	OnBeforeEnterFn func()
	OnEnterFn       func()
	OnExitFn        func()
	OnProcessFn     func(duration time.Duration)
	ExitCond        func() bool
	StageType       StageType
	TimeToLive      time.Duration

	subStage int32
	enterAt  time.Time
	exitAt   time.Time

	age time.Duration
}

func (stage *StageInfo) OnProcess(delta time.Duration) {
	if stage.OnProcessFn != nil {
		stage.OnProcessFn(delta)
	}
	stage.age += delta
}

func (stage *StageInfo) OnBeforeEnter() {
	if stage.OnBeforeEnterFn != nil {
		stage.OnBeforeEnterFn()
	}
	stage.subStage = 1
}

func (stage *StageInfo) OnEnter() {
	if stage.OnEnterFn != nil {
		stage.OnEnterFn()
	}
	stage.subStage = 10
	stage.enterAt = time.Now()
}

func (stage *StageInfo) OnExit() {
	if stage.OnExitFn != nil {
		stage.OnExitFn()
	}
	stage.subStage = 100
	stage.exitAt = time.Now()
}

func (stage *StageInfo) CheckExit() bool {
	if stage.Timeout() {
		return true
	}

	if stage.ExitCond != nil {
		return stage.ExitCond()
	}
	return true
}

func (stage *StageInfo) Timeout() bool {
	if stage.TimeToLive > 0 {
		return stage.age > stage.TimeToLive
	}
	return false
}
