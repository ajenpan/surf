package cli

import (
	"bytes"
	"fmt"
	"runtime"
)

var DefaultVersion = NewAppInfo()

func NewAppInfo() *AppInfo {
	return &AppInfo{
		Name:       "unknow",
		Version:    "unknow",
		GitCommit:  "unknow",
		BuildAt:    "unknow",
		BuildBy:    runtime.Version(),
		RunnningOS: runtime.GOOS + "/" + runtime.GOARCH,
	}
}

type AppInfo struct {
	Name       string
	Version    string
	GitCommit  string
	BuildAt    string
	BuildBy    string
	RunnningOS string
}

func (info *AppInfo) ShortVersion() string {
	return info.Version
}

func (info *AppInfo) LongVersion() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintln(buf, "project:", info.Name)
	fmt.Fprintln(buf, "version:", info.Version)
	fmt.Fprintln(buf, "git commit:", info.GitCommit)
	fmt.Fprintln(buf, "build at:", info.BuildAt)
	fmt.Fprintln(buf, "build by:", info.BuildBy)
	fmt.Fprintln(buf, "running OS/Arch:", info.RunnningOS)
	return buf.String()
}
