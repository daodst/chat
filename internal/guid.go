package internal

import (
	"encoding/json"
	"git.9885.net/golib/util/ajwt"
)

type JwtParam struct {
	Uid      string `json:"uid"`      // Id
	Username string `json:"username"` 
	Type     int    `json:"type"`     // token，1access_token,2refresh_token
	Device   string `json:"device"`   // ，，consts/device.go

	Appid string `json:"appid"` 
	Uuid  int64  `json:"uuid"`  

	Dat int64 `json:"dat"` // jwt，
}

func DecryptInfo(guid string, keyStr string) (res []byte, err error) {
	
	info := &JwtParam{}
	err = ajwt.RsaDecryptObj(guid, &info, keyStr)
	if err != nil {
		return
	}
	res, err = json.Marshal(info)
	return
}
