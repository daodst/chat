package auth

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"

	"freemasonry.cc/chat/clientapi/userutil"
	"strings"

	"freemasonry.cc/chat/clientapi/auth/authtypes"
	"freemasonry.cc/chat/clientapi/httputil"
	"freemasonry.cc/chat/clientapi/jsonerror"
	"freemasonry.cc/chat/setup/config"
	uapi "freemasonry.cc/chat/userapi/api"
	"github.com/matrix-org/util"
	"net/http"
)

// LoginTypeXs describes how to authenticate with a login token.
type LoginTypeXs struct {
	UserAPI uapi.LoginTokenInternalAPI
	Config  *config.ClientAPI
}

// Name implements Type.
func (t *LoginTypeXs) Name() string {
	return authtypes.LoginTypeXs
}

// LoginFromJSON implements Type. The cleanup function deletes the token from
// the database on success.
func (t *LoginTypeXs) LoginFromJSON(ctx context.Context, reqBytes []byte) (*Login, LoginCleanupFunc, *util.JSONResponse) {
	var r loginXsRequest
	if err := httputil.UnmarshalJSON(reqBytes, &r); err != nil {
		return nil, nil, err
	}
	login, err := t.Login(ctx, &r)
	if err != nil {
		return nil, nil, err
	}

	return login, func(context.Context, *util.JSONResponse) {}, nil
}

func (t *LoginTypeXs) Login(ctx context.Context, req interface{}) (*Login, *util.JSONResponse) {
	r := req.(*loginXsRequest)
	username := strings.ToLower(r.Username())
	if username == "" {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("A username must be supplied."),
		}
	}
	msgBytes := crypto.Keccak256([]byte(r.Timestamp))
	sigBytes, err := hex.DecodeString(r.Sign)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("！"),
		}
	}
	//TODO ，
	// username
	pubKey, err := crypto.SigToPub(msgBytes, sigBytes)
	if err != nil || pubKey == nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("！"),
		}
	}
	addressHex := crypto.PubkeyToAddress(*pubKey).Hex()[2:]
	recoveredAddress := strings.ToLower(addressHex)

	localpart, err := userutil.ParseUsernameParam(username, &t.Config.Matrix.ServerName)
	if err != nil {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.InvalidUsername(err.Error()),
		}
	}
	if localpart != recoveredAddress {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("！"),
		}
	}
	
	recoverPubKeyBytes := crypto.CompressPubkey(pubKey)
	//reqPubKeyBytes, err := hex.DecodeString(r.PubKey)
	//if err != nil || !bytes.Equal(recoverPubKeyBytes, reqPubKeyBytes) {
	//	return nil, &util.JSONResponse{
	//		Code: http.StatusUnauthorized,
	//		JSON: jsonerror.BadJSON("！"),
	//	}
	//}
	if !crypto.VerifySignature(recoverPubKeyBytes, msgBytes, sigBytes[:len(sigBytes)-1]) {
		return nil, &util.JSONResponse{
			Code: http.StatusUnauthorized,
			JSON: jsonerror.BadJSON("！"),
		}
	}
	return &r.Login, nil
}

// loginXsRequest struct to hold the possible parameters from an HTTP request.
type loginXsRequest struct {
	Login
	Token     string `json:"token"`
	Sign      string `json:"sign"`
	Timestamp string `json:"timestamp"`
	PubKey    string `json:"pub_key"`
}
