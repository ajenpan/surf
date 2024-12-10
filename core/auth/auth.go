package auth

import (
	"encoding/json"
	"fmt"
)

type UserInfo struct {
	UId   uint32 `json:"uid"`
	URole uint16 `json:"urid"`
}

func (u *UserInfo) UserID() uint32 {
	return u.UId
}
func (u *UserInfo) UserRole() uint16 {
	return u.URole
}

type NodeInfo struct {
	NId      uint32 `json:"nid"`
	NType    uint16 `json:"ntype"`
	NName    string `json:"nname"`
	NVersion string `json:"nversion"`
}

func (n *NodeInfo) NodeID() uint32 {
	return n.NId
}

func (n *NodeInfo) NodeName() string {
	return n.NName
}

func (n *NodeInfo) String() string {
	return fmt.Sprintf("node://%d/%d", n.NId, n.NType)
}

func (n *NodeInfo) NodeType() uint16 {
	return n.NType
}

func (n *NodeInfo) NodeVersion() string {
	return n.NVersion
}

func (n *NodeInfo) UserID() uint32 {
	return n.NId
}
func (n *NodeInfo) UserName() string {
	return n.NodeName()
}
func (n *NodeInfo) UserRole() uint16 {
	return n.NType
}

func (n *NodeInfo) Marshal() []byte {
	v, _ := json.Marshal(n)
	return v
}

func (n *NodeInfo) Unmarshal(data []byte) error {
	return json.Unmarshal(data, n)
}
