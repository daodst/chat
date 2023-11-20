package new_feature

import (
	"errors"
	"fmt"
	"freemasonry.cc/chat/new_feature/new_db"
	"time"
)

func CleanExpireHistoryMsgs(interval int64) error {
	parseDuration, _ := time.ParseDuration(fmt.Sprintf("-%dh", interval*24))
	expireTime := time.Now().Add(parseDuration)
	roomIds := new_db.GetAllRoomIds()
	for _, roomId := range roomIds {
		ok := new_db.DeleteExpireEventsByRoomId(roomId, expireTime)
		if !ok {
			return errors.New("err when cleaning room:" + roomId)
		}
	}
	return nil
}
