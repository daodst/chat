package routing

import (
	"encoding/hex"
	"errors"
	"freemasonry.cc/chat/clientapi/httputil"
	"freemasonry.cc/chat/clientapi/jsonerror"
	"freemasonry.cc/chat/internal/eventutil"
	"freemasonry.cc/chat/new_feature"
	"freemasonry.cc/chat/new_feature/new_db"
	roomserverAPI "freemasonry.cc/chat/roomserver/api"
	"freemasonry.cc/chat/setup/config"
	"freemasonry.cc/chat/syncapi/sync"
	userapi "freemasonry.cc/chat/userapi/api"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/matrix-org/util"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"net/http"
	"strings"
	"time"
)

type queryReq struct {
	PubKey    string `json:"pub_key"`
	QuerySign string `json:"query_sign"`
	Invitee   string `json:"invitee"`
	Timestamp string `json:"timestamp"`
	Localpart string `json:"localpart"`
}

var errMissingUserID = errors.New("'user_id' must be supplied")

func QueryInviteRes(
	httpReq *http.Request,
) util.JSONResponse {
	// 1, ，，chat_addr
	req := queryReq{}
	resErr := httputil.UnmarshalJSONRequest(httpReq, &req)
	if resErr != nil {
		return *resErr
	}
	if req.QuerySign == "" || req.Invitee == "" || req.PubKey == "" {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.BadJSON("invalid params"),
		}
	}
	// 1.1, 
	pubKeyBytes, err := hex.DecodeString(req.PubKey)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.Unknown("err hex fmt of PubKey"),
		}
	}
	signBytes, err := hex.DecodeString(req.QuerySign)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.Unknown("err hex fmt of QuerySign"),
		}
	}
	msgBytes := []byte(req.Timestamp)
	pubKeyChat := ethsecp256k1.PubKey{Key: pubKeyBytes}
	addressHexChat := sdk.AccAddress(pubKeyChat.Address()).String()
	//addressHex := pubKey.Address().String()
	recoveredChatAddress := strings.ToLower(addressHexChat)
	
	if !pubKeyChat.VerifySignature(msgBytes, signBytes) {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.BadJSON("check sign err"),
		}
	}
	// 2, chat_addrlocalpart ：1，  2，，，
	localPart, err := new_feature.GetLocalByChatAddr(recoveredChatAddress)
	if err != nil {
		util.GetLogger(httpReq.Context()).WithError(err).Error("new_feature.GetLocalByChatAddr")
	}
	if localPart == "" {
		localPart = req.Localpart
	}
	// 3, ，get/user/by/phone
	res, err := new_feature.QueryInviteRes(localPart, req.Invitee)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: jsonerror.Unknown(err.Error()),
		}
	}

	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: res,
	}
}

type inviteAsOwnerReq struct {
	RoomId       string `json:"room_id"`
	TargetUserId string `json:"target_user_id"`
	Reason       string `json:"reason"`
}

// UserInactive3Days 
func UserInactive3Days(httpReq *http.Request) util.JSONResponse {
	//ctx := httpReq.Context()
	req := []string{}
	res := []string{}
	resErr := httputil.UnmarshalJSONRequest(httpReq, &req)
	if resErr != nil {
		return *resErr
	}
	if len(req) == 0 {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.BadJSON("invalid params"),
		}
	}

	for i, _ := range req {
		if value, ok := sync.SoPool.PositiveUsers[req[i]]; ok {
			if value>>1|0b000 == 0 { 
				res = append(res, req[i])
			}
		}
	}

	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: res,
	}
}

// UserInactive7Days 
func UserInactive7Days(httpReq *http.Request) util.JSONResponse {
	//ctx := httpReq.Context()
	req := []string{}
	res := []string{}
	resErr := httputil.UnmarshalJSONRequest(httpReq, &req)
	if resErr != nil {
		return *resErr
	}
	if len(req) == 0 {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.BadJSON("invalid params"),
		}
	}

	for i, _ := range req {
		if value, ok := sync.SoPool.PositiveUsers[req[i]]; ok {
			if value>>1|0b0000000 == 0 { 
				res = append(res, req[i])
			}
		}
	}

	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: res,
	}
}

func SendOwnerInvite(httpReq *http.Request, userAPI userapi.FederationUserAPI, cfg *config.FederationAPI, rsAPI roomserverAPI.FederationRoomserverAPI,
) util.JSONResponse {
	ctx := httpReq.Context()
	req := inviteAsOwnerReq{}
	resErr := httputil.UnmarshalJSONRequest(httpReq, &req)
	if resErr != nil {
		return *resErr
	}
	if req.RoomId == "" || req.TargetUserId == "" {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.BadJSON("invalid params"),
		}
	}
	groupOwner, err := new_db.GetGroupOwner(req.RoomId)
	if err != nil {
		return util.ErrorResponse(err)
	}
	var userPro userapi.QueryProfileResponse
	err = userAPI.QueryProfile(ctx, &userapi.QueryProfileRequest{
		req.TargetUserId,
	}, &userPro)
	if err != nil {
		return util.ErrorResponse(err)
	}
	builder := gomatrixserverlib.EventBuilder{
		Sender:   groupOwner,
		RoomID:   req.RoomId,
		Type:     "m.room.member",
		StateKey: &req.TargetUserId,
	}
	content := gomatrixserverlib.MemberContent{
		Membership:  gomatrixserverlib.Invite,
		DisplayName: userPro.DisplayName,
		AvatarURL:   userPro.AvatarURL,
		Reason:      req.Reason,
		IsDirect:    false,
	}
	if err = builder.SetContent(content); err != nil {
		return util.ErrorResponse(err)
	}
	event, err := eventutil.QueryAndBuildEvent(ctx, &builder, cfg.Matrix, time.Now(), rsAPI, nil)
	if err == errMissingUserID {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.BadJSON(err.Error()),
		}
	} else if err == eventutil.ErrRoomNoExists {
		return util.JSONResponse{
			Code: http.StatusNotFound,
			JSON: jsonerror.NotFound(err.Error()),
		}
	} else if err != nil {
		util.GetLogger(ctx).WithError(err).Error("buildMembershipEvent failed")
		return jsonerror.InternalServerError()
	}

	var inviteRes roomserverAPI.PerformInviteResponse
	if err := rsAPI.PerformInvite(ctx, &roomserverAPI.PerformInviteRequest{ 
		Event:           event,
		InviteRoomState: nil, // ask the roomserver to draw up invite room state for us
		RoomVersion:     event.RoomVersion,
		SendAsServer:    string(cfg.Matrix.ServerName),
	}, &inviteRes); err != nil {
		util.GetLogger(ctx).WithError(err).Error("PerformInvite failed")
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: jsonerror.InternalServerError(),
		}
	}
	if inviteRes.Error != nil {
		return inviteRes.Error.JSONResponse()
	}
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: struct{}{},
	}
}
