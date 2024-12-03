package core

import (
	"bytes"
	"fmt"
	"runtime"
)

type Server interface {
	ServerType() uint16
	ServerName() string
}

type ServerInfo struct {
	Name       string
	Version    string
	GitCommit  string
	BuildAt    string
	BuildBy    string
	RunnningOS string
}

func NewServerInfo() *ServerInfo {
	return &ServerInfo{
		BuildBy:    runtime.Version(),
		RunnningOS: runtime.GOOS + "/" + runtime.GOARCH,
	}
}

func (s *ServerInfo) LongVersion() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintln(buf, "project:", s.Name)
	fmt.Fprintln(buf, "version:", s.Version)
	fmt.Fprintln(buf, "git commit:", s.GitCommit)
	fmt.Fprintln(buf, "build at:", s.BuildAt)
	fmt.Fprintln(buf, "build by:", s.BuildBy)
	fmt.Fprintln(buf, "running OS/Arch:", s.RunnningOS)
	return buf.String()
}
