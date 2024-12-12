package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type AuthClient struct {
	RemoteUrl string
}

func (ac *AuthClient) TokenAuth(token []byte) (*UserInfo, error) {
	resp, err := http.Post(ac.RemoteUrl, "application/json", bytes.NewReader(token))
	if err != nil {
		return nil, err
	}
	ret := &UserInfo{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(ret)
	return ret, err
}

func NewAuthClientFunc(remote string) func([]byte) (*UserInfo, error) {
	ac := &AuthClient{RemoteUrl: remote}
	return ac.TokenAuth
}
