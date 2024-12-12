package game

import (
	"fmt"
	"sync"
)

var LogicCreator = &GameLogicCreator{}

func RegisterGame(name string, creator func() Logic) error {
	return LogicCreator.Add(name, creator)
}

type GameLogicCreator struct {
	Store sync.Map
}

func (c *GameLogicCreator) Add(name string, creator func() Logic) error {
	c.Store.Store(name, creator)
	return nil
}

func (c *GameLogicCreator) CreateLogic(name string) (Logic, error) {
	v, has := c.Store.Load(name)
	if !has {
		return nil, fmt.Errorf("game logic %s not found", name)
	}
	creator := v.(func() Logic)
	return creator(), nil
}