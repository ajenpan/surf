package core

type NodeType = uint16

const (
	NodeType_Client NodeType = 0

	NodeType_Core  NodeType = 100
	NodeType_Gate  NodeType = 101
	NodeType_UAuth NodeType = 102

	NodeType_Lobby  NodeType = 103
	NodeType_Battle NodeType = 104
)
