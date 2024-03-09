package media_manage

import (
	"freemasonry.cc/chat/new_feature/media_manage/disk"
	"os"
)

func GetMaxPanSize() uint64 {
	path, _ := os.Getwd()
	di, _ := disk.GetInfo(path)
	return di.Free
}
