package media_manage

import (
	"fmt"
	"freemasonry.cc/blockchain/client"
	"freemasonry.cc/chat/new_feature"
	"freemasonry.cc/chat/new_feature/new_db"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const UserBasicSize = 1024 * 1024 * 60

var (
	txClient   = client.NewTxClient()
	accClient  = client.NewAccountClient(&txClient)
	chatClient = client.NewChatClient(&txClient, &accClient)
)

type MediaManager struct {
	lock          sync.RWMutex
	MediaFlowMap  map[string]*MediaFlowInfo
	MaxPanSize    uint64
	CantAvailable bool
	CleanInterval int64 // clean the file timeout for several days
}

func NewMediaManager(cleanInterval int64) *MediaManager {
	manager := &MediaManager{
		MediaFlowMap:  make(map[string]*MediaFlowInfo, 64),
		MaxPanSize:    GetMaxPanSize() / 2 / uint64(cleanInterval),
		CleanInterval: cleanInterval,
	}
	go manager.Run()
	return manager
}

// GetMediaFlowInfo fetches the NetflowInfo matching the addr
func (m *MediaManager) GetMediaFlowInfo(addr string) (*MediaFlowInfo, bool) {
	m.lock.RLock()
	info, exist := m.MediaFlowMap[addr]
	m.lock.RUnlock()
	return info, exist
}

// MediaFlowInfoCount returns the number of existing allocations
func (m *MediaManager) MediaFlowInfoCount() int {
	res := len(m.MediaFlowMap)
	return res
}

// CreateMediaFlowInfo get or create a MediaFlowInfo user by username
func (m *MediaManager) CreateMediaFlowInfo(addr string) *MediaFlowInfo {
	if info, exist := m.GetMediaFlowInfo(addr); exist {
		return info
	} else {
		
		infoNew := NewMediaFlowInfo(addr, UserBasicSize)
		m.lock.Lock()
		m.MediaFlowMap[addr] = infoNew
		m.lock.Unlock()
		m.UpdateMediaFlowInfo(addr)
		m.UpdateAllMediaFlowInfo()
		return infoNew
	}
}

func (m *MediaManager) UpdateAllMediaFlowInfo() {
	// ，NetflowInfo
	m.lock.Lock()
	for _, info := range m.MediaFlowMap {
		limit := m.CalcMediaFlowLimit(info)
		info.FlowLimit = int64(limit)
	}
	m.lock.Unlock()
}

func (m *MediaManager) CleanMediaFlowInfo() {
	// ，NetflowInfo
	m.lock.RLock()
	m.MediaFlowMap = make(map[string]*MediaFlowInfo)
	m.lock.RUnlock()
}

func (m *MediaManager) UpdateMediaFlowInfo(addr string) {
	
	m.lock.RLock()
	info, ok := m.MediaFlowMap[addr]
	m.lock.RUnlock()

	if !ok {
		infoNew := NewMediaFlowInfo(addr, UserBasicSize)
		m.lock.Lock()
		m.MediaFlowMap[addr] = infoNew
		m.lock.Unlock()
		info = infoNew
	}
	info.UserImprove = m.GetChainInfoByAddr(addr)
}

func (m *MediaManager) CalcMediaFlowLimit(info *MediaFlowInfo) int {
	
	basicTmp := float64(int(m.MaxPanSize) / m.MediaFlowInfoCount())
	return int(basicTmp * info.UserImprove)
}

func (m *MediaManager) GetChainInfoByAddr(addr string) float64 {
	rate := 1.00
	chatGain, err := chatClient.QueryChatGain(addr)
	if err != nil {
		return rate
	} else {
		rate100 := chatGain.Int64() + 100
		rate = float64(rate100) / 100.00
	}
	return rate
}

func (m *MediaManager) CleanTimeoutFiles() {
	pathThis, _ := os.Getwd()
	parseDuration, _ := time.ParseDuration(fmt.Sprintf("-%dh", m.CleanInterval*24))
	ts := time.Now().Add(parseDuration).UnixMilli()
	base64Hashs := new_db.GetTimeoutFilesBase64Hash(ts)
	if len(base64Hashs) > 0 {
		for _, base64Hash := range base64Hashs {
			filePathTmp := filepath.Join(pathThis, "media_store", base64Hash[0:1], base64Hash[1:2], base64Hash[2:])
			os.RemoveAll(filePathTmp)
		}
		new_db.DeleteMediaRepository(base64Hashs)
	}
}

func (m *MediaManager) CleanFilesForUser(userId string, size int64) int64 { // remove some files for user to save a new file
	pathThis, _ := os.Getwd()
	MetaL := new_db.GetFilesBase64HashForUser(userId)
	var sum int64 = 0
	for _, meta := range MetaL {
		exUse := new_db.GetFilesBase64HashExUse(userId, meta.MediaId)
		if len(exUse) == 0 {
			filePathTmp := filepath.Join(pathThis, "media_store", meta.Base64hash[0:1], meta.Base64hash[1:2], meta.Base64hash[2:])
			os.RemoveAll(filePathTmp)
		}
		new_db.DeleteMediaRepositoryById(meta.MediaId)
		sum += meta.FileSizeBytes
		if sum >= size {
			break
		}
	}
	return sum
}

func (m *MediaManager) UpdateMaxPanSize() {
	m.lock.Lock()
	m.MaxPanSize = GetMaxPanSize() / 2 / uint64(m.CleanInterval)
	m.lock.Unlock()
	m.UpdateAllMediaFlowInfo()
}

func (m *MediaManager) Run() {
	timerCleanTimeoutFiles := time.NewTicker(time.Hour * 24) //  * time.Duration(m.CleanInterval)
	timerCleanNetflowInfo := time.NewTicker(time.Hour * 24)
	timerUpdateMaxPanSize := time.NewTicker(time.Hour * 24)
	for {
		select {
		case <-timerCleanTimeoutFiles.C:
			m.CleanTimeoutFiles()
			err := new_feature.CleanExpireHistoryMsgs(m.CleanInterval)
			if err != nil {
				log.WithError(err).Error("new_feature.CleanExpireHistoryMsgs")
			}
		case <-timerCleanNetflowInfo.C:
			m.CleanMediaFlowInfo()
		case <-timerUpdateMaxPanSize.C:
			m.UpdateMaxPanSize()

		}
	}
}
