package media_manage

type MediaFlowInfo struct {
	Username    string  `json:"username"`      
	CurUsedFlow int64   `json:"cur_used_flow"` 
	FlowLimit   int64   `json:"flow_limit"`    
	UserImprove float64 `json:"user_improve"`  
}

func NewMediaFlowInfo(username string, flowCouldPerDay int64) *MediaFlowInfo {
	return &MediaFlowInfo{
		Username:  username,
		FlowLimit: flowCouldPerDay,
	}
}

func (m *MediaFlowInfo) InPacket(packetLen int64) {
	m.CurUsedFlow += packetLen
}
