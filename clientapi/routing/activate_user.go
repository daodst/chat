package routing

import (
	"context"
	"freemasonry.cc/chat/clientapi/jsonerror"
	"freemasonry.cc/chat/new_feature/new_db"
	"freemasonry.cc/chat/setup/config"
	"freemasonry.cc/chat/syncapi/sync"
	"freemasonry.cc/chat/userapi/api"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/matrix-org/util"
	"net/http"
)

func GetInactiveUsersByRoom(req *http.Request, device *api.Device, roomId string,
	cfg *config.ClientAPI,
	federation *gomatrixserverlib.FederationClient) util.JSONResponse {
	// ID
	owner, err2 := new_db.GetGroupOwner(roomId)
	if err2 != nil {
		return util.ErrorResponse(err2)
	}
	if owner != device.UserID {
		err3 := jsonerror.Unknown("no permission")
		return util.JSONResponse{
			Code: http.StatusForbidden,
			JSON: err3,
		}
	}
	users, err := new_db.UsersJoinedRoomId(roomId)
	if err != nil {
		return util.ErrorResponse(err)
	}
	//overTwoDays, inactiveFor3Days, err := GetUserOffline2DayOrInactive3Day(req.Context(), federation, users, string(cfg.Matrix.ServerName))
	//if err != nil {
	//	return util.ErrorResponse(err)
	//}
	overThreeDays, inactiveFor3Days, err2 := GetUserOffline3DayOrInactive3Day(req.Context(), federation, users, string(cfg.Matrix.ServerName))
	if err2 != nil {
		return util.ErrorResponse(err2)
	}
	overSevenDays, inactiveFor7Days, err2 := GetUserOffline3DayOrInactive3Day(req.Context(), federation, users, string(cfg.Matrix.ServerName))
	if err2 != nil {
		return util.ErrorResponse(err2)
	}
	res := make(map[string][]string)
	res["offline_3_days"] = overThreeDays
	res["inactive_3_days"] = inactiveFor3Days
	res["offline_7_days"] = overSevenDays
	res["inactive_7_days"] = inactiveFor7Days
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: res,
	}
}
func GetUserOffline2DayOrInactive3Day(ctx context.Context, federation *gomatrixserverlib.FederationClient, userIds []string, server string) (overTwoDays []string, inactiveFor3Days []string, err error) {
	var serversMap = make(map[string][]string)
	for i, _ := range userIds {
		// serverusers
		_, domain, _ := gomatrixserverlib.SplitID('@', userIds[i])
		if _, ok := serversMap[string(domain)]; !ok {
			serversMap[string(domain)] = []string{userIds[i]}
		} else {
			serversMap[string(domain)] = append(serversMap[string(domain)], userIds[i])
		}
	}
	res, err := new_db.GetUsersOfflineOver2Days(userIds)
	if err != nil {
		util.GetLogger(ctx).Error(err)
	}
	if len(res) > 0 {
		overTwoDays = append(overTwoDays, res...)
	}
	//	map， TODOwaitgroup，
	for s, _ := range serversMap {
		//	server， federation
		if s == server {
			//	
			for i, _ := range serversMap[s] {
				if value, ok := sync.SoPool.PositiveUsers[serversMap[s][i]]; ok {
					if value>>1|0b000 == 0 { 
						inactiveFor3Days = append(inactiveFor3Days, serversMap[s][i])
					}
				}
			}
		} else {
			//	
			// todo presence
			inactives, err := federation.Inactive3Days(ctx, gomatrixserverlib.ServerName(s), serversMap[s])
			if err != nil {
				util.GetLogger(ctx).Error(err)
				continue
			}
			inactiveFor3Days = append(inactiveFor3Days, inactives...)
		}
	}
	
	return
}

func GetUserOffline3DayOrInactive3Day(ctx context.Context, federation *gomatrixserverlib.FederationClient, userIds []string, server string) (overThreeDays []string, inactiveFor3Days []string, err error) {
	var serversMap = make(map[string][]string)
	for i, _ := range userIds {
		// serverusers
		_, domain, _ := gomatrixserverlib.SplitID('@', userIds[i])
		if _, ok := serversMap[string(domain)]; !ok {
			serversMap[string(domain)] = []string{userIds[i]}
		} else {
			serversMap[string(domain)] = append(serversMap[string(domain)], userIds[i])
		}
	}
	res, err := new_db.GetUsersOfflineOver3Days(userIds)
	if err != nil {
		util.GetLogger(ctx).Error(err)
	}
	if len(res) > 0 {
		overThreeDays = append(overThreeDays, res...)
	}
	//	map， TODOwaitgroup，
	for s, _ := range serversMap {
		//	server， federation
		if s == server {
			//	
			for i, _ := range serversMap[s] {
				if value, ok := sync.SoPool.PositiveUsers[serversMap[s][i]]; ok {
					if value>>1|0b000 == 0 { 
						inactiveFor3Days = append(inactiveFor3Days, serversMap[s][i])
					}
				}
			}
		} else {
			//	
			// todo presence
			inactives, err := federation.Inactive3Days(ctx, gomatrixserverlib.ServerName(s), serversMap[s])
			if err != nil {
				util.GetLogger(ctx).Error(err)
				continue
			}
			inactiveFor3Days = append(inactiveFor3Days, inactives...)
		}
	}
	
	return
}

func GetUserOffline7DayOrInactive7Day(ctx context.Context, federation *gomatrixserverlib.FederationClient, userIds []string, server string) (overSevenDays []string, inactiveFor7Days []string, err error) {
	var serversMap = make(map[string][]string)
	for i, _ := range userIds {
		// serverusers
		_, domain, _ := gomatrixserverlib.SplitID('@', userIds[i])
		if _, ok := serversMap[string(domain)]; !ok {
			serversMap[string(domain)] = []string{userIds[i]}
		} else {
			serversMap[string(domain)] = append(serversMap[string(domain)], userIds[i])
		}
	}
	res, err := new_db.GetUsersOfflineOver7Days(userIds)
	if err != nil {
		util.GetLogger(ctx).Error(err)
	}
	if len(res) > 0 {
		overSevenDays = append(overSevenDays, res...)
	}
	//	map， TODOwaitgroup，
	for s, _ := range serversMap {
		//	server， federation
		if s == server {
			//	
			for i, _ := range serversMap[s] {
				if value, ok := sync.SoPool.PositiveUsers[serversMap[s][i]]; ok {
					if value>>1|0b0000000 == 0 { 
						inactiveFor7Days = append(inactiveFor7Days, serversMap[s][i])
					}
				}
			}
		} else {
			//	
			// todo presence
			inactives, err := federation.Inactive7Days(ctx, gomatrixserverlib.ServerName(s), serversMap[s])
			if err != nil {
				util.GetLogger(ctx).Error(err)
				continue
			}
			inactiveFor7Days = append(inactiveFor7Days, inactives...)
		}
	}
	
	return
}
