package server

import "github.com/ajenpan/surf/core"

const (
	NodeType_UAuth  core.NodeType = 102
	NodeType_Lobby  core.NodeType = 103
	NodeType_Battle core.NodeType = 104
)

const (
	NodeName_Client string = "client"
	NodeName_Core   string = "core"
	NodeName_Gate   string = "gate"
	NodeName_UAuth  string = "uauth"
	NodeName_Lobby  string = "lobby"
	NodeName_Battle string = "battle"
)

func NodeName(ntype core.NodeType) string {
	switch ntype {
	case core.NodeType_Client:
		return NodeName_Client
	case core.NodeType_Core:
		return NodeName_Core
	case core.NodeType_Gate:
		return NodeName_Gate
	case NodeType_UAuth:
		return NodeName_UAuth
	case NodeType_Lobby:
		return NodeName_Lobby
	case NodeType_Battle:
		return NodeName_Battle
	}
	return ""
}
