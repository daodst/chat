package routing

import (
	"context"
	"errors"
	"freemasonry.cc/chat/clientapi/httputil"
	"freemasonry.cc/chat/clientapi/jsonerror"
	"freemasonry.cc/chat/roomserver/api"
	"freemasonry.cc/chat/roomserver/version"
	"freemasonry.cc/chat/setup/config"
	userapi "freemasonry.cc/chat/userapi/api"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/matrix-org/gomatrixserverlib/tokens"
	"github.com/matrix-org/util"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type sendGroupMasterNoticeRequest struct {
	UserID  string `json:"user_id,omitempty"`
	Content struct {
		MsgType string `json:"msgtype,omitempty"`
		Body    string `json:"body,omitempty"`
	} `json:"content,omitempty"`
	Type     string `json:"type,omitempty"`
	StateKey string `json:"state_key,omitempty"`
	RoomId   string `json:"room_id,omitempty"`
}

func (r sendGroupMasterNoticeRequest) valid() bool {
	// todo 
	return true
}
func SendGroupMasterNotice(
	req *http.Request,
	cfgClient *config.ClientAPI,
	userAPI userapi.ClientUserAPI,
	rsAPI api.ClientRoomserverAPI,
) util.JSONResponse {
	ctx := req.Context()

	var r sendGroupMasterNoticeRequest
	resErr := httputil.UnmarshalJSONRequest(req, &r) // req
	if resErr != nil {
		return *resErr
	}
	senderDevice, err := getSenderDeviceForSpecUser(ctx, r.UserID, userAPI, cfgClient)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.Unknown("err getSenderDeviceForSpecUser"),
		}
	}
	// check that all required fields are set  
	if !r.valid() {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.BadJSON("Invalid request"),
		}
	}

	var (
		roomID      = r.RoomId // todo roomID 
		roomVersion = version.DefaultRoomVersion()
	)
	// todo APP，
	request := map[string]interface{}{
		"body":    r.Content.Body,
		"msgtype": r.Content.MsgType,
	}
	e, resErr := generateSendEvent(ctx, request, senderDevice, roomID, "m.room.message", nil, cfgClient, rsAPI, time.Now())
	if resErr != nil {
		logrus.Errorf("failed to send message: %+v", resErr)
		return *resErr
	}

	var txnAndSessionID *api.TransactionID

	// pass the new event to the roomserver and receive the correct event ID
	// event ID in case of duplicate transaction is discarded
	if err := api.SendEvents(
		ctx, rsAPI,
		api.KindNew,
		[]*gomatrixserverlib.HeaderedEvent{
			e.Headered(roomVersion),
		},
		cfgClient.Matrix.ServerName,
		cfgClient.Matrix.ServerName,
		txnAndSessionID,
		false,
	); err != nil {
		util.GetLogger(ctx).WithError(err).Error("SendEvents failed")
		return jsonerror.InternalServerError()
	}
	util.GetLogger(ctx).WithFields(logrus.Fields{
		"event_id":     e.EventID(),
		"room_id":      roomID,
		"room_version": roomVersion,
	}).Info("Sent event to roomserver")

	res := util.JSONResponse{
		Code: http.StatusOK,
		JSON: sendEventResponse{e.EventID()},
	}

	return res
}

// getSenderDeviceForSpecUser
// It returns an userapi.Device, which is used for building the event
func getSenderDeviceForSpecUser(
	ctx context.Context,
	userId string,
	userAPI userapi.ClientUserAPI,
	cfg *config.ClientAPI,
) (*userapi.Device, error) {
	// Check if we got existing devices
	deviceRes := &userapi.QueryDevicesResponse{}
	err := userAPI.QueryDevices(ctx, &userapi.QueryDevicesRequest{
		UserID: userId,
	}, deviceRes)
	if err != nil {
		return nil, err
	}

	if len(deviceRes.Devices) > 0 {
		return &deviceRes.Devices[0], nil
	}
	local, domain, err := gomatrixserverlib.SplitID('@', userId)
	if err != nil {
		return nil, err
	}
	if domain != cfg.Matrix.ServerName {
		return nil, errors.New("wrong servername")
	}
	// create an AccessToken
	token, err := tokens.GenerateLoginToken(tokens.TokenOptions{
		ServerPrivateKey: cfg.Matrix.PrivateKey.Seed(),
		ServerName:       string(cfg.Matrix.ServerName),
		UserID:           userId,
	})
	if err != nil {
		return nil, err
	}

	// create a new device, if we didn't find any
	var devRes userapi.PerformDeviceCreationResponse
	err = userAPI.PerformDeviceCreation(ctx, &userapi.PerformDeviceCreationRequest{
		Localpart:          local,
		DeviceDisplayName:  &local,
		AccessToken:        token,
		NoDeviceListUpdate: true,
	}, &devRes)

	if err != nil {
		return nil, err
	}
	return devRes.Device, nil
}
