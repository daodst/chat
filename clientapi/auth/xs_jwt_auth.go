// Copyright 2021 The Matrix.org Foundation C.I.C.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"freemasonry.cc/chat/clientapi/userutil"
	"freemasonry.cc/chat/internal"
	"net/http"
	"strings"

	"freemasonry.cc/chat/clientapi/auth/authtypes"
	"freemasonry.cc/chat/clientapi/httputil"
	"freemasonry.cc/chat/clientapi/jsonerror"
	"freemasonry.cc/chat/setup/config"
	uapi "freemasonry.cc/chat/userapi/api"
	"github.com/matrix-org/util"
)

// LoginTypeToken describes how to authenticate with a login token.
type LoginTypeJwt struct {
	UserAPI uapi.ClientUserAPI
	Config  *config.ClientAPI
}

// Name implements Type.
func (t *LoginTypeJwt) Name() string {
	return authtypes.LoginTypeJwt
}

// LoginFromJSON implements Type. The cleanup function deletes the token from
// the database on success.
func (t *LoginTypeJwt) LoginFromJSON(ctx context.Context, reqBytes []byte) (*Login, LoginCleanupFunc, *util.JSONResponse) {
	var r loginJwtRequest
	if err := httputil.UnmarshalJSON(reqBytes, &r); err != nil {
		return nil, nil, err
	}

	login, err := t.Login(ctx, &r)
	if err != nil {
		return nil, nil, err
	}

	return login, func(context.Context, *util.JSONResponse) {}, nil
}

func (t *LoginTypeJwt) Login(ctx context.Context, req interface{}) (*Login, *util.JSONResponse) {
	r := req.(*loginJwtRequest)
	username := strings.ToLower(r.Username())
	if username == "" || r.Jwt == "" {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.MissingParam("A username and jwt must be supplied."),
		}
	}
	localpart, err := userutil.ParseUsernameParam(username, &t.Config.Matrix.ServerName)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.InvalidUsername(err.Error()),
		}
	}
	// jwt
	//2，jwt，jwtuidusername
	got, err := internal.DecryptInfo(r.Jwt, t.Config.Matrix.GuidPubKey)
	if err != nil {
		util.GetLogger(ctx).WithField("r.Jwt", r.Jwt).WithField("t.Config.Matrix.GuidPubKey", t.Config.Matrix.GuidPubKey).WithError(err).Error("DecryptInfo")
		return nil, &util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: jsonerror.Unknown("failed to decrypt jwt:" + err.Error()),
		}
	}
	gotStruct := internal.JwtParam{}
	err = json.Unmarshal(got, &gotStruct)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: jsonerror.Unknown("failed to Unmarshal:" + err.Error()),
		}
	}
	if localpart != gotStruct.Uid {
		return nil, &util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: jsonerror.Forbidden("username and jwt mismatch"),
		}
	}
	
	res := &uapi.QueryAccountAvailabilityResponse{}
	err = t.UserAPI.QueryAccountAvailability(ctx, &uapi.QueryAccountAvailabilityRequest{
		Localpart: localpart,
	}, res)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: jsonerror.Unknown("failed to check availability:" + err.Error()),
		}
	}
	if res.Available {
		res1 := &uapi.PerformAccountCreationResponse{}
		
		err := t.UserAPI.PerformAccountCreation(ctx, &uapi.PerformAccountCreationRequest{
			AccountType: uapi.AccountTypeUser,
			Localpart:   username,
			OnConflict:  uapi.ConflictAbort,
		}, res1)
		if err != nil {
			return nil, &util.JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: jsonerror.Unknown("failed to create account:" + err.Error()),
			}
		}
		// api
		err = t.UserAPI.InsertChainData(ctx, username, string(t.Config.Matrix.ServerName), "", "")
		if err != nil {
			return nil, &util.JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: jsonerror.Unknown("failed to InsertChainData when creating account:" + err.Error()),
			}
		}
		if gotStruct.Uuid != 0 {
			err = t.UserAPI.AddTelNumbers(ctx, username, []string{fmt.Sprintf("%d", gotStruct.Uuid)})
			if err != nil {
				return nil, &util.JSONResponse{
					Code: http.StatusInternalServerError,
					JSON: jsonerror.Unknown("failed to AddTelNumbers when creating account:" + err.Error()),
				}
			}
		}
	}

	return &r.Login, nil
}

// loginTokenRequest struct to hold the possible parameters from an HTTP request.
type loginJwtRequest struct {
	Login
	Jwt string `json:"jwt"`
}
