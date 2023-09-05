package routing

import (
	appserviceAPI "github.com/matrix-org/dendrite/appservice/api"
	"github.com/matrix-org/dendrite/setup/config"
	userapi "github.com/matrix-org/dendrite/userapi/api"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/matrix-org/util"
	"net/http"
)

// GetHostByServerNode implements GET /get/host/of/{nodeName}
func GetHostByServerNode(
	req *http.Request, profileAPI userapi.ClientUserAPI, cfg *config.ClientAPI,
	nodeName string, asAPI appserviceAPI.AppServiceInternalAPI,
	federation *gomatrixserverlib.FederationClient,
) util.JSONResponse {
	host := gomatrixserverlib.GetHostByNode(nodeName)

	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: host,
	}
}
