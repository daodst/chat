package routing

import (
	appserviceAPI "github.com/matrix-org/dendrite/appservice/api"
	"github.com/matrix-org/dendrite/clientapi/httputil"
	"github.com/matrix-org/dendrite/clientapi/jsonerror"
	"github.com/matrix-org/dendrite/new_feature"
	"github.com/matrix-org/dendrite/setup/config"
	"github.com/matrix-org/dendrite/userapi/api"
	userapi "github.com/matrix-org/dendrite/userapi/api"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/matrix-org/util"
	"net/http"
)

type QueryAvailableReq struct {
	Locals []string `json:"locals"`
}

// QueryAvailableUsers implements /new/query_available_users/by/locals
func QueryAvailableUsers(req *http.Request, device *api.Device) util.JSONResponse {
	r := QueryAvailableReq{}
	if resErr := httputil.UnmarshalJSONRequest(req, &r); resErr != nil {
		return *resErr
	}
	if len(r.Locals) == 0 {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.BadJSON("'locals' must be supplied."),
		}
	}
	res := make(map[string]interface{})
	localpart, _, err := gomatrixserverlib.SplitID('@', device.UserID)
	if err != nil {
		util.GetLogger(req.Context()).WithError(err).Error("gomatrixserverlib.SplitID failed")
		return jsonerror.InternalServerError()
	}
	availableList, needPayList, cantList, err := new_feature.QueryAvailableByLocals(localpart, r.Locals)

	res["available_list"] = availableList

	res["need_pay_list"] = needPayList
	res["cant_list"] = cantList
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: res,
	}
}

type GetUserByPhoneReq struct {
	Phone int64 `json:"phone"`
}

func GetUserByPhone(req *http.Request, device *api.Device, profileAPI userapi.ClientUserAPI, cfg *config.ClientAPI,
	asAPI appserviceAPI.AppServiceInternalAPI,
	federation *gomatrixserverlib.FederationClient) util.JSONResponse {
	r := GetUserByPhoneReq{}
	if resErr := httputil.UnmarshalJSONRequest(req, &r); resErr != nil {
		return *resErr
	}
	if r.Phone == 0 {
		return util.JSONResponse{
			Code: http.StatusBadRequest,
			JSON: jsonerror.BadJSON("'phone' must be supplied."),
		}
	}
	localpart, _, err := gomatrixserverlib.SplitID('@', device.UserID)
	if err != nil {
		util.GetLogger(req.Context()).WithError(err).Error("gomatrixserverlib.SplitID failed")
		return jsonerror.InternalServerError()
	}
	res, err := new_feature.GetUserByPhone(localpart, r.Phone)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: jsonerror.Unknown(err.Error()),
		}
	}

	userId := "@" + res.Localpart + ":" + res.Servername
	profile, err := getProfile(req.Context(), profileAPI, cfg, userId, asAPI, federation)
	if err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: jsonerror.Unknown(err.Error()),
		}
	}
	res.DisplayName = profile.DisplayName
	res.AvatarURL = profile.AvatarURL

	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: res,
	}
}
