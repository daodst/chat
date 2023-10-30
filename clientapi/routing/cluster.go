package routing

import (
	"context"
	appserviceAPI "freemasonry.cc/chat/appservice/api"
	"freemasonry.cc/chat/clientapi/jsonerror"
	"freemasonry.cc/chat/internal/eventutil"
	"freemasonry.cc/chat/new_feature"
	"freemasonry.cc/chat/new_feature/new_db"
	roomserverAPI "freemasonry.cc/chat/roomserver/api"
	"freemasonry.cc/chat/setup/config"
	"freemasonry.cc/chat/userapi/api"
	userapi "freemasonry.cc/chat/userapi/api"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/matrix-org/util"
	"net/http"
	"time"
)

// RejoinCluster 
func RejoinCluster(req *http.Request, device *api.Device, roomId string,
	cfg *config.ClientAPI,
	profileAPI api.ClientUserAPI, rsAPI roomserverAPI.ClientRoomserverAPI,
	asAPI appserviceAPI.AppServiceInternalAPI, federation *gomatrixserverlib.FederationClient) util.JSONResponse {
	local, _, err := gomatrixserverlib.SplitID('@', device.UserID)
	if err != nil {
		return util.ErrorResponse(err)
	}
	
	isinEquipmentGroup, err := new_feature.JudgeIsinEquipmentGroup(roomId, local)
	if err != nil {
		return util.ErrorResponse(err)
	}
	if isinEquipmentGroup {
		
		conflictUsers, err := new_db.GetConflictUsers(device.UserID, roomId)
		if err != nil {
			util.GetLogger(req.Context()).Error(err)
		}
		if len(conflictUsers) > 0 {
			for i, _ := range conflictUsers { 
				SendLeaveAsRoomOwner(req.Context(), profileAPI, roomId, conflictUsers[i], "cluster_addr_conflict", cfg, rsAPI, asAPI, time.Now())
			}
		}
		
		_, domain, err := gomatrixserverlib.SplitID('!', roomId)
		if err != nil {
			return util.ErrorResponse(err)
		}
		if domain == cfg.Matrix.ServerName {
			
			return SendInviteAsRoomOwner(req.Context(), profileAPI, roomId, device.UserID, "from_cluster", cfg, rsAPI, asAPI, time.Now())
		} else {
			// ，federationapiinvite
			
			res, err := federation.SendInviteAsOwner(req.Context(), domain, make(map[string]interface{}))
			if err != nil {
				return util.ErrorResponse(err)
			}
			return util.JSONResponse{
				Code: http.StatusOK,
				JSON: res,
			}
		}
	}
	return util.JSONResponse{
		Code: http.StatusForbidden,
		JSON: jsonerror.Unknown("not in the cluster"),
	}
}
func SendInviteAsRoomOwner(ctx context.Context,
	profileAPI userapi.ClientUserAPI,
	roomID, userID, reason string,
	cfg *config.ClientAPI,
	rsAPI roomserverAPI.ClientRoomserverAPI,
	asAPI appserviceAPI.AppServiceInternalAPI, evTime time.Time) util.JSONResponse {

	owner, err := new_db.GetGroupOwner(roomID)
	if err != nil {
		return util.ErrorResponse(err)
	}
	profile, err := loadProfile(ctx, userID, cfg, profileAPI, asAPI)
	if err != nil {
		return util.ErrorResponse(err)
	}

	builder := gomatrixserverlib.EventBuilder{
		Sender:   owner,
		RoomID:   roomID,
		Type:     "m.room.member",
		StateKey: &userID,
	}

	content := gomatrixserverlib.MemberContent{
		Membership:  gomatrixserverlib.Invite,
		DisplayName: profile.DisplayName,
		AvatarURL:   profile.AvatarURL,
		Reason:      reason,
		IsDirect:    false,
	}

	if err = builder.SetContent(content); err != nil {
		return util.ErrorResponse(err)
	}
	event, err := eventutil.QueryAndBuildEvent(ctx, &builder, cfg.Matrix, evTime, rsAPI, nil)
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

// SendLeaveAsRoomOwner 
func SendLeaveAsRoomOwner(ctx context.Context,
	profileAPI userapi.ClientUserAPI,
	roomID, userID, reason string,
	cfg *config.ClientAPI,
	rsAPI roomserverAPI.ClientRoomserverAPI,
	asAPI appserviceAPI.AppServiceInternalAPI, evTime time.Time) util.JSONResponse {

	owner, err := new_db.GetGroupOwner(roomID)
	if err != nil {
		return util.ErrorResponse(err)
	}
	profile, err := loadProfile(ctx, userID, cfg, profileAPI, asAPI)
	if err != nil {
		return util.ErrorResponse(err)
	}

	builder := gomatrixserverlib.EventBuilder{
		Sender:   owner,
		RoomID:   roomID,
		Type:     "m.room.member",
		StateKey: &userID,
	}

	content := gomatrixserverlib.MemberContent{
		Membership:  gomatrixserverlib.Leave,
		DisplayName: profile.DisplayName,
		AvatarURL:   profile.AvatarURL,
		Reason:      reason,
		IsDirect:    false,
	}

	if err = builder.SetContent(content); err != nil {
		return util.ErrorResponse(err)
	}
	event, err := eventutil.QueryAndBuildEvent(ctx, &builder, cfg.Matrix, evTime, rsAPI, nil)
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

	if err = roomserverAPI.SendEvents(
		ctx, rsAPI,
		roomserverAPI.KindNew,
		[]*gomatrixserverlib.HeaderedEvent{event.Event.Headered(event.RoomVersion)},
		cfg.Matrix.ServerName,
		cfg.Matrix.ServerName,
		nil,
		false,
	); err != nil {
		util.GetLogger(ctx).WithError(err).Error("SendEvents failed")
		return jsonerror.InternalServerError()
	}

	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: struct{}{},
	}
}
