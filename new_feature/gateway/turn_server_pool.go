package gateway

import (
	"context"
	"freemasonry.cc/blockchain/client"
	"freemasonry.cc/blockchain/x/gateway/types"
	"github.com/matrix-org/util"
	"time"
)

var (
	err               error
	logger            = util.GetLogger(context.Background())
	txClient          = client.NewTxClient()
	gatewayClient     = client.NewGatewayClinet(&txClient)
	gatewayList       []types.GatewayListResp
	updateGatewayFunc func()
)

func init() {
	updateGatewayFunc = func() {
		gatewayList, err = gatewayClient.QueryGatewayList()
		if err != nil {
			logger.WithError(err).Error("gatewayClient.QueryGatewayList()")
		}
		time.AfterFunc(time.Hour, updateGatewayFunc)
	}
	time.AfterFunc(time.Minute, updateGatewayFunc)

}
