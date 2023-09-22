// Copyright 2017 Michael Telatysnki <7t3chguy@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/matrix-org/dendrite/internal"
	"github.com/matrix-org/dendrite/new_feature"
	"github.com/matrix-org/gomatrixserverlib"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/matrix-org/dendrite/clientapi/jsonerror"
	"github.com/matrix-org/dendrite/setup/config"
	"github.com/matrix-org/dendrite/userapi/api"
	userapi "github.com/matrix-org/dendrite/userapi/api"
	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/util"
)

var turnHost = "127.0.0.1"

// RequestTurnServer implements:

//	GET /voip/turnServer
func RequestTurnServer(profileAPI userapi.ClientUserAPI, req *http.Request, device *api.Device, cfg *config.ClientAPI) util.JSONResponse {
	logger := util.GetLogger(req.Context())
	turnConfig := cfg.TURN
	// TODO Guest Support
	if len(turnConfig.URIs) == 0 || turnConfig.UserLifetime == "" {
		return util.JSONResponse{
			Code: http.StatusOK,
			JSON: struct{}{},
		}
	}
	turnHost = strings.Split(turnConfig.URIs[0], ":")[1]
	// Duration checked at startup, err not possible
	duration, _ := time.ParseDuration(turnConfig.UserLifetime)

	resp := gomatrix.RespTurnServer{
		URIs: turnConfig.URIs,
		TTL:  int(duration.Seconds()),
	}

	if turnConfig.SharedSecret != "" {
		expiry := time.Now().Add(duration).Unix()
		resp.Username = fmt.Sprintf("%d:%s", expiry, device.UserID)
		mac := hmac.New(sha1.New, []byte(turnConfig.SharedSecret))
		_, err := mac.Write([]byte(resp.Username))

		if err != nil {
			logger.WithError(err).Error("mac.Write failed")
			return jsonerror.InternalServerError()
		}

		resp.Password = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	} else if turnConfig.Username != "" && turnConfig.Password != "" {
		//resp.Username = turnConfig.Username
		//resp.Password = turnConfig.Password
		
		resp.Username = device.UserID
		resp.Password = util.RandomString(8)
		// limit  ，，，
		limit := int64(5)
		if cfg.Matrix.Mode == "chain" {
			localpart, _, err := gomatrixserverlib.SplitID('@', device.UserID)
			if err != nil {
				util.GetLogger(req.Context()).WithError(err).Error("gomatrixserverlib.SplitID failed")
				return jsonerror.InternalServerError()
			}

			res := new_feature.QueryUserInfoByLocal(localpart)
			limit, err = internal.CalcLimitByLevel(res.MortgageLevel)
			if err != nil {
				return util.JSONResponse{
					Code: http.StatusInternalServerError,
					JSON: jsonerror.Unknown("CalcLimit"),
				}
			}
		}

		apiClient := getHttpClient(time.Second * 5)
		defer apiClient.CloseIdleConnections()
		reqHttp := turnReq{}
		reqHttp.Username = device.UserID
		reqHttp.Realm = "dendrite"
		reqHttp.Password = resp.Password
		reqHttp.Limit = limit

		respHttp := turnReq{}
		turnUrlTmp := fmt.Sprintf("http://%s:23478/turn/addTmpAuth", turnHost)
		if err := PostJSON(req.Context(), &apiClient, turnUrlTmp, reqHttp, &respHttp); err != nil {
			logger.WithError(err).WithField("reqHttp", reqHttp).WithField("turnUrlTmp", turnUrlTmp).Error("err addTmpAuth")
			return util.JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: jsonerror.Unknown("addTmpAuth"),
			}
		}
	} else {
		return util.JSONResponse{
			Code: http.StatusOK,
			JSON: struct{}{},
		}
	}

	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: resp,
	}
}

type turnReq struct {
	Username string `json:"username"`
	Realm    string `json:"realm"`
	Password string `json:"password"`
	Limit    int64  `json:"limit"`
}

// AddTmpAuthInfo implements:

//	POST /voip/turnAuthTmpAdd
func AddTmpAuthInfo(profileAPI userapi.ClientUserAPI, req *http.Request, device *api.Device, cfg *config.ClientAPI) util.JSONResponse {
	turnConfig := cfg.TURN
	// Duration checked at startup, err not possible
	if len(turnConfig.URIs) == 0 || turnConfig.UserLifetime == "" {
		return jsonerror.InternalServerError()
	}
	turnHost = strings.Split(turnConfig.URIs[0], ":")[1]
	duration, _ := time.ParseDuration(turnConfig.UserLifetime)
	resp := gomatrix.RespTurnServer{
		URIs: turnConfig.URIs,
		TTL:  int(duration.Seconds()),
	}
	resp.Username = device.UserID
	resp.Password = util.RandomString(8)

	// limit  ，limit
	limit := int64(5)
	if cfg.Matrix.Mode == "chain" {
		res := &userapi.QueryProfileResponse{}
		err := profileAPI.QueryProfile(req.Context(), &userapi.QueryProfileRequest{UserID: device.UserID}, res)
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: jsonerror.Unknown("QueryProfile"),
			}
		}
		limit, err = internal.CalcLimit(res.MortgageFee)
		if err != nil {
			return util.JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: jsonerror.Unknown("CalcLimit"),
			}
		}
	}
	apiClient := getHttpClient(time.Second * 5)
	defer apiClient.CloseIdleConnections()
	reqHttp := turnReq{}
	reqHttp.Username = device.UserID
	reqHttp.Realm = "dendrite"
	reqHttp.Password = resp.Password
	reqHttp.Limit = limit

	respHttp := turnReq{}

	if err := PostJSON(req.Context(), &apiClient, fmt.Sprintf("http://%s:23478/turn/addTmpAuth", turnHost), reqHttp, &respHttp); err != nil {
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: jsonerror.Unknown("PostJSON"),
		}
	}
	
	//if err := turn_server.AddTmpAuth(resp.Username, "dendrite", resp.Password, 5); err != nil {
	//	return jsonerror.InternalServerError()
	//}

	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: resp,
	}
}

// DelTmpAuthInfo implements:

//	POST /voip/turnAuthTmpDel
func DelTmpAuthInfo(req *http.Request, device *api.Device, cfg *config.ClientAPI) util.JSONResponse {
	apiClient := getHttpClient(time.Second * 5)
	defer apiClient.CloseIdleConnections()
	reqHttp := turnReq{}
	reqHttp.Username = device.UserID

	respHttp := turnReq{}

	if err := PostJSON(req.Context(), &apiClient, fmt.Sprintf("http://%s:23478/turn/delTmpAuth", turnHost), reqHttp, &respHttp); err != nil {
		return jsonerror.InternalServerError()
	}
	//if err := turn_server.DelTmpAuth(device.UserID); err != nil {
	//	return jsonerror.InternalServerError()
	//}
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: struct {
		}{},
	}
}

func getHttpClient(options ...interface{}) http.Client {
	timeout0 := time.Minute * 1
	if len(options) > 0 {
		timeout0 = options[0].(time.Duration)
	}
	return http.Client{
		Timeout: timeout0,
	}
}

// PostJSON performs a POST request with JSON on an internal HTTP API
func PostJSON(
	ctx context.Context, httpClient *http.Client,
	apiURL string, request, response interface{},
) error {
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	parsedAPIURL, err := url.Parse(apiURL)
	if err != nil {
		return err
	}

	parsedAPIURL.Path = strings.TrimLeft(parsedAPIURL.Path, "/")
	apiURL = parsedAPIURL.String()

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := httpClient.Do(req.WithContext(ctx))
	if res != nil {
		defer (func() { err = res.Body.Close() })()
	}
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		var errorBody struct {
			Message string `json:"message"`
		}
		if msgerr := json.NewDecoder(res.Body).Decode(&errorBody); msgerr == nil {
			return fmt.Errorf("internal API: %d from %s: %s", res.StatusCode, apiURL, errorBody.Message)
		}
		return fmt.Errorf("internal API: %d from %s", res.StatusCode, apiURL)
	}
	return json.NewDecoder(res.Body).Decode(response)
}
