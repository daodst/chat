package auth

import (
	"context"
	"encoding/hex"
	_ "freemasonry.cc/blockchain/client"
	util2 "freemasonry.cc/blockchain/util"
	"freemasonry.cc/chat/clientapi/auth/authtypes"
	"freemasonry.cc/chat/clientapi/httputil"
	"freemasonry.cc/chat/clientapi/jsonerror"
	"freemasonry.cc/chat/clientapi/userutil"
	"freemasonry.cc/chat/new_feature"
	"freemasonry.cc/chat/new_feature/new_db"
	"freemasonry.cc/chat/setup/config"
	uapi "freemasonry.cc/chat/userapi/api"
	"github.com/matrix-org/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"net/http"
	"os"
	"strings"
)

var AmtRegisterUsers = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "dendrite",
		Subsystem: "client",
		Name:      "register_users_total",
		Help:      "Total number of register users on this node",
	},
)

// LoginTypeCosmos describes how to authenticate with a login token.
type LoginTypeCosmos struct {
	UserAPI uapi.ClientUserAPI
	Config  *config.ClientAPI
}

// Name implements Type.
func (t *LoginTypeCosmos) Name() string {
	return authtypes.LoginTypeCosmos
}

// LoginFromJSON implements Type. The cleanup function deletes the token from
// the database on success.
func (t *LoginTypeCosmos) LoginFromJSON(ctx context.Context, reqBytes []byte) (*Login, LoginCleanupFunc, *util.JSONResponse) {
	var r loginCosmosRequest
	if err := httputil.UnmarshalJSON(reqBytes, &r); err != nil {
		return nil, nil, err
	}
	login, err := t.Login(ctx, &r)
	if err != nil {
		return nil, nil, err
	}

	return login, func(context.Context, *util.JSONResponse) {}, nil
}

func (t *LoginTypeCosmos) Login(ctx context.Context, req interface{}) (*Login, *util.JSONResponse) {

	r := req.(*loginCosmosRequest)
	username := strings.ToLower(r.Username())
	if username == "" {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("A username must be supplied."),
		}
	}
	localpart, err := userutil.ParseUsernameParam(username, &t.Config.Matrix.ServerName)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.InvalidUsername(err.Error()),
		}
	}
	if os.Getenv("CHAT_SERVER_MODE") == "chain" {
		_, err = new_feature.GetUserByLocal(localpart)
		if err == nil { // err == nil ，
			
			servernameOk := new_feature.CheckServername(localpart, string(t.Config.Matrix.ServerName))
			if !servernameOk {
				return nil, &util.JSONResponse{
					Code: http.StatusUnauthorized,
					JSON: jsonerror.Unknown("err servername"),
				}
			}
		}
	}
	msgBytes := []byte(r.Timestamp)
	sigBytes, err := hex.DecodeString(r.Sign)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("err hex decode"),
		}
	}
	sigChatBytes, err := hex.DecodeString(r.ChatSign)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("err hex decode"),
		}
	}
	//TODO ，
	// username
	pubKeyBytes, err := hex.DecodeString(r.PubKey)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("err recovering pub"),
		}
	}
	pubKeyChatBytes, err := hex.DecodeString(r.ChatPubKey)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("err recovering chat pub"),
		}
	}
	pubKey := ethsecp256k1.PubKey{Key: pubKeyBytes}
	recoveredAddress, err := util2.GetAccountFromPub(r.PubKey)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("err GetAccountFromPub"),
		}
	}
	//cfg := sdk.GetConfig()
	//cmdCfg.SetBech32Prefixes(cfg)
	//cmdCfg.SetBip44CoinType(cfg)
	//addressHex := sdk.AccAddress(pubKey.Address()).String()
	//addressHex := pubKey.Address().String()
	//recoveredAddress := strings.ToLower(addressHex)
	pubKeyChat := ethsecp256k1.PubKey{Key: pubKeyChatBytes}

	//addressHexChat := sdk.AccAddress(pubKeyChat.Address()).String()
	//addressHex := pubKey.Address().String()
	//recoveredChatAddress := strings.ToLower(addressHexChat)
	recoveredChatAddress, err := util2.GetAccountFromPub(r.ChatPubKey)
	if localpart != recoveredAddress {
		util.GetLogger(ctx).Errorf("localpart=%+v recoveredAddress=%v", localpart, recoveredAddress)
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("crypto address mismatch with recovered"),
		}
	}
	
	if !pubKey.VerifySignature(msgBytes, sigBytes) {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("check sign err"),
		}
	}
	
	if !pubKeyChat.VerifySignature(msgBytes, sigChatBytes) {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("check chat sign err"),
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
		} else {
			AmtRegisterUsers.Inc()
			
			err2 := new_feature.SendServerNotice(ctx, res1.Account.UserID, "u.welcome_register", "u.welcome_register", "m.room.message")
			if err2 != nil {

			}
		}
		
		if err = new_db.UpsertChatBindAddr(username, recoveredChatAddress); err != nil {
			return nil, &util.JSONResponse{
				Code: http.StatusInternalServerError,
				JSON: jsonerror.Unknown("failed to set chat_addr:" + err.Error()),
			}
		}
		//// api  
		//err = t.UserAPI.InsertChainData(ctx, username, string(t.Config.Matrix.ServerName), "", "")
		//if err != nil {
		//	return nil, &util.JSONResponse{
		//		Code: http.StatusInternalServerError,
		//		JSON: jsonerror.Unknown("failed to InsertChainData when creating account:" + err.Error()),
		//	}
		//}
	}
	return &r.Login, nil
}

// loginCosmosRequest struct to hold the possible parameters from an HTTP request.
type loginCosmosRequest struct {
	Login
	Token      string `json:"token"`
	Sign       string `json:"sign"`
	Timestamp  string `json:"timestamp"`
	PubKey     string `json:"pub_key"`
	ChatPubKey string `json:"chat_pub_key"`
	ChatSign   string `json:"chat_sign"`
}
