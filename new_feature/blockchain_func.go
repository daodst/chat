package new_feature

import (
	"context"
	"errors"
	"fmt"
	"freemasonry.cc/blockchain/client"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/matrix-org/util"
	"os"
	"strings"
)

var (
	logger     = util.GetLogger(context.Background())
	txClient   = client.NewTxClient()
	accClient  = client.NewAccountClient(&txClient)
	chatClient = client.NewChatClient(&txClient, &accClient)
)

type InfoRes struct {
	Localpart     string   `json:"localpart"`
	LimitMode     string   `json:"limit_mode"`
	ChatFee       string   `json:"chat_fee"`
	Servername    string   `json:"servername"`
	Blacklist     []string `json:"blacklist"`
	Whitelist     []string `json:"whitelist"`
	MortgageLevel int64    `json:"mortgage_level"`
	TelNumbers    []string `json:"tel_numbers"`
}

func QueryUserInfoByLocal(local string) InfoRes {
	res := InfoRes{}
	info, err := chatClient.QueryUserInfo(local)
	if err != nil {
		return res
	}
	if info.Status != 0 {
		return res
	}
	resInfo := info.UserInfo
	res = InfoRes{
		Localpart:     resInfo.FromAddress,
		LimitMode:     resInfo.ChatRestrictedMode,
		ChatFee:       resInfo.ChatFee.Amount.String(),
		Servername:    resInfo.GatewayProfixMobile + ".fm",
		Blacklist:     resInfo.ChatBlacklist,
		Whitelist:     resInfo.ChatWhitelist,
		MortgageLevel: resInfo.PledgeLevel,
		TelNumbers:    resInfo.Mobile,
	}
	return res
}

func JudgeIfPayByLocals(from, to string) bool {
	logger.WithField("from", from).WithField("to", to).Info("JudgeIfPayByLocals")
	isPay, err := chatClient.QueryChatSendGift(from, to)
	if err != nil {
		logger.WithError(err).Error("err JudgeIfPayByLocals")
		//return false   
	}

	return isPay
}

func GetPayRelationByLocals(from string, to []string) (map[string]bool, error) {
	return chatClient.QueryChatSendGifts(from, to)
}

type InviteRes struct {
	Localpart     string   `json:"localpart"`
	LimitMode     string   `json:"limit_mode"`
	ChatFee       string   `json:"chat_fee"`
	Servername    string   `json:"servername"`
	Blacklist     []string `json:"blacklist"`
	Whitelist     []string `json:"whitelist"`
	Payed         bool     `json:"payed"`
	Reason        string   `json:"reason"`
	MortgageLevel int64    `json:"mortgage_level"`
	TelNumbers    []string `json:"tel_numbers"`
}

func QueryAvailableByLocals(userLocal string, locals []string) (available, needPays, cantChat []InviteRes, err error) {
	available, needPays, cantChat = []InviteRes{}, []InviteRes{}, []InviteRes{}
	userInfos, err := chatClient.QueryUserInfos(locals)
	if err != nil {
		return
	}
	payRelationByLocals, err := GetPayRelationByLocals(userLocal, locals)
	if err != nil {
		return
	}
	for _, info := range userInfos {
		if info.IsExist == 0 {
			itemInfo := InviteRes{
				Localpart:     info.FromAddress,
				LimitMode:     info.ChatRestrictedMode,
				ChatFee:       info.ChatFee.Amount.String(),
				Servername:    info.GatewayProfixMobile + ".fm",
				Blacklist:     info.ChatBlacklist,
				Whitelist:     info.ChatWhitelist,
				MortgageLevel: info.PledgeLevel,
				TelNumbers:    info.Mobile,
			}
			if b, ok := payRelationByLocals[info.FromAddress]; ok {
				itemInfo.Payed = b
			}
			resTmp, reason := CheckAvailable(userLocal, itemInfo)
			itemInfo.Blacklist = []string{}
			itemInfo.Whitelist = []string{}
			switch resTmp {
			case "ok":
				available = append(available, itemInfo)
			case "pay":
				itemInfo.Reason = reason
				needPays = append(needPays, itemInfo)
			default:
				itemInfo.Reason = reason
				cantChat = append(cantChat, itemInfo)
			}
		} else {
			itemInfo := InviteRes{
				Localpart: info.FromAddress,
				Reason:    "not_exist_on_chain",
			}
			cantChat = append(cantChat, itemInfo)
		}
	}
	return
}

func CheckAvailable(userLocal string, info InviteRes) (res, reason string) {
	blackListStr := strings.Join(info.Blacklist, ",")
	if strings.Contains(blackListStr, userLocal) {
		return "cant", "in_blacklist"
	}
	whiteListStr := strings.Join(info.Whitelist, ",")
	if strings.Contains(whiteListStr, userLocal) {
		return "ok", ""
	}
	switch info.LimitMode {
	case "list":
		if !strings.Contains(whiteListStr, userLocal) {
			return "cant", "out_of_whitelist"
		}
	case "fee":
		if info.Payed {
			return "ok", ""
		} else {
			return "pay", "need_pay"
		}
	case "any":
		return "ok", ""
	}
	return "ok", ""
}

func CheckMortgage(userLocal string) bool {
	pledgeInfo, err2 := chatClient.QueryPledgeInfo(userLocal)
	if err2 != nil {
		logger.WithError(err2).Error("chatClient.QueryPledgeInfo")
		return false
	}
	need, _ := types.NewIntFromString("100000000000000000000")

	return pledgeInfo.AllPledgeAmount.Amount.GTE(need)
}

func JudgeShouldAutoJoin(inviterLocal, inviteeLocal string) (bool, error) {
	if os.Getenv("CHAT_SERVER_MODE") != "chain" {
		return true, nil
	}
	inviteeInfo, err := chatClient.QueryUserInfo(inviteeLocal)
	if err != nil {
		return false, err
	}
	if inviteeInfo.Status != 1 {
		return false, errors.New(inviteeInfo.Message)
	} else {
		limitMode := "fee"
		if inviteeInfo.UserInfo.ChatRestrictedMode != "" {
			limitMode = inviteeInfo.UserInfo.ChatRestrictedMode
		}
		whiteListStr := strings.Join(inviteeInfo.UserInfo.ChatWhitelist, ",")
		blackListStr := strings.Join(inviteeInfo.UserInfo.ChatBlacklist, ",")
		if strings.Contains(blackListStr, inviterLocal) {
			return false, nil
		}
		if strings.Contains(whiteListStr, inviterLocal) {
			return true, nil
		}
		switch limitMode {
		case "fee":
			return inviteeInfo.UserInfo.ChatFee.IsNil() || JudgeIfPayByLocals(inviterLocal, inviteeLocal), err // inviterRelation.PresentFee != ""API，center
			// can talk only if inviter has sent gift before,inviteeSettings.ChatFee == "" 
		case "list":
			return strings.Contains(whiteListStr, inviterLocal) && !strings.Contains(blackListStr, inviterLocal), nil
		case "any":
			return true, nil
		}
	}

	return false, nil
}

type UserRes struct {
	DisplayName string `json:"display_name"` 
	AvatarURL   string `json:"avatar_url"`   

	Localpart     string `json:"localpart"`      
	Servername    string `json:"servername"`     
	LimitMode     string `json:"limit_mode"`     //  any fee list
	ChatFee       string `json:"chat_fee"`       
	MortgageFee   string `json:"mortgage_fee"`   
	MortgageLevel int64  `json:"mortgage_level"` 

	TelNumbers []string `json:"tel_numbers"` 
	Blacklist  []string `json:"blacklist"`   
	Whitelist  []string `json:"whitelist"`   

	CanWeTalk  bool   `json:"can_we_talk"`  
	CanPayTalk bool   `json:"can_pay_talk"` 
	PayedFee   string `json:"payed_fee"`    
	Payed      bool   `json:"payed"`        
}

// GetUserByPhone 
func GetUserByPhone(userLocal string, phone int64) (res UserRes, err error) {
	userInfo, err := chatClient.QueryUserByMobile(fmt.Sprintf("%d", phone))
	if err != nil {
		return
	}
	res = UserRes{
		DisplayName:   userInfo.FromAddress,
		AvatarURL:     "",
		Localpart:     userInfo.FromAddress,
		Servername:    userInfo.GatewayProfixMobile + ".fm",
		LimitMode:     userInfo.ChatRestrictedMode,
		ChatFee:       userInfo.ChatFee.Amount.String(),
		MortgageFee:   "",
		MortgageLevel: userInfo.PledgeLevel,
		TelNumbers:    userInfo.Mobile,
		Blacklist:     userInfo.ChatBlacklist,
		Whitelist:     userInfo.ChatWhitelist,
	}
	switch res.LimitMode {
	case "any":
		res.CanWeTalk = true
	case "fee":
		if JudgeIfPayByLocals(userLocal, res.Localpart) {
			res.Payed = true
			res.CanWeTalk = true
		} else {
			res.CanPayTalk = true
		}
	case "list":

	}
	whiteListStr := strings.Join(res.Whitelist, ",")
	if strings.Contains(whiteListStr, userLocal) {
		res.CanWeTalk = true
	}

	return
}
