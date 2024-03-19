package media_manage

import (
	"fmt"
	"freemasonry.cc/chat/new_feature/new_db"
	"sync"
	"testing"
	"time"
)

func TestMediaManager(t *testing.T) {
	users := []struct {
		Local string
	}{
		{"aaa"},
		{"bbb"},
		{"ccc"},
	}
	t.Log(" --- NewMediaManager --- ")
	mediaManager := NewMediaManager(3)
	t.Logf("MediaManager: %v", mediaManager)
	t.Log(" --- simulate upload --- ")
	for _, user := range users {
		go UserUseMediaFlow(user.Local, mediaManager, t)
	}
	select {}
}

func UserUseMediaFlow(local string, manager *MediaManager, t *testing.T) {
	uMediaFlow := manager.CreateMediaFlowInfo(local)
	i := 0
	for {
		if i%20 == 0 {
			t.Log(fmt.Sprintf("user: %s, total:%dM, cur_used: %dM", uMediaFlow.Username, uMediaFlow.FlowLimit/1024/1024, uMediaFlow.CurUsedFlow/1024/1024))
		}
		var inputFilesize int64 = 10 * 1024 * 1024

		if uMediaFlow.FlowLimit-uMediaFlow.CurUsedFlow < inputFilesize {
			if inputFilesize > uMediaFlow.FlowLimit {
				t.Log(local, "over limit")
			}
			//  ，over_limit  
			sizeTmp := inputFilesize - (uMediaFlow.FlowLimit - uMediaFlow.CurUsedFlow)
			//mediaManager.CleanFilesForUser(string(r.MediaMetadata.UserID), sizeTmp)  // 。log
			t.Log(local, "clean some space for file upload:", sizeTmp)
			uMediaFlow.CurUsedFlow -= sizeTmp
			uMediaFlow.InPacket(sizeTmp)
			t.Log(local, "InPacket:", sizeTmp)
		} else {
			uMediaFlow.InPacket(inputFilesize)
			t.Log(local, "InPacket:", inputFilesize)
		}
		i++
		time.Sleep(time.Second * 2)
	}
}

func TestMediaManager_CleanTimeoutFiles(t *testing.T) {
	new_db.InitDb("postgresql://postgres:postgres@127.0.0.1:5432/chat?sslmode=disable")
	type fields struct {
		lock          sync.RWMutex
		MediaFlowMap  map[string]*MediaFlowInfo
		MaxPanSize    uint64
		CantAvailable bool
		CleanInterval int64
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "1", fields: fields{
			MediaFlowMap:  make(map[string]*MediaFlowInfo),
			CleanInterval: 90,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MediaManager{
				lock:          tt.fields.lock,
				MediaFlowMap:  tt.fields.MediaFlowMap,
				MaxPanSize:    tt.fields.MaxPanSize,
				CantAvailable: tt.fields.CantAvailable,
				CleanInterval: tt.fields.CleanInterval,
			}
			m.CleanTimeoutFiles()
		})
	}
}
