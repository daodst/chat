package new_feature

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func GetHttpClient(options ...interface{}) http.Client {
	timeout0 := time.Minute * 1
	if len(options) > 0 {
		timeout0 = options[0].(time.Duration)
	}
	return http.Client{
		Timeout: timeout0,
	}
}

// PostJSONNoSpan performs a POST request with JSON on an internal HTTP API
func PostJSONNoSpan(
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

type sendNoticeRequest struct {
	UserID  string `json:"user_id,omitempty"`
	Content struct {
		MsgType string `json:"msgtype,omitempty"`
		Body    string `json:"body,omitempty"`
	} `json:"content,omitempty"`
	Type     string `json:"type,omitempty"`
	StateKey string `json:"state_key,omitempty"`
	RoomId   string `json:"room_id,omitempty"`
}

func SendServerNotice(ctx context.Context, userId, msgType, body, outType string) error {
	apiClient := GetHttpClient(time.Minute * 1)
	defer apiClient.CloseIdleConnections()
	req := sendNoticeRequest{
		UserID: userId,
		Content: struct {
			MsgType string `json:"msgtype,omitempty"`
			Body    string `json:"body,omitempty"`
		}(struct {
			MsgType string
			Body    string
		}{MsgType: msgType, Body: body}),
		Type:     outType,
		StateKey: "",
	}
	res := struct {
		EventId string `json:"event_id,omitempty"`
	}{}
	return PostJSONNoSpan(ctx, &apiClient, LocalServerUrl+ServerNoticePath, req, &res)
}

func SendRoomNotice(ctx context.Context, userId, msgType, body, outType, roomId string) error {
	apiClient := GetHttpClient(time.Minute * 1)
	defer apiClient.CloseIdleConnections()
	req := sendNoticeRequest{
		UserID: userId,
		Content: struct {
			MsgType string `json:"msgtype,omitempty"`
			Body    string `json:"body,omitempty"`
		}(struct {
			MsgType string
			Body    string
		}{MsgType: msgType, Body: body}),
		Type:     outType,
		StateKey: "",
		RoomId:   roomId,
	}
	res := struct {
		EventId string `json:"event_id,omitempty"`
	}{}
	return PostJSONNoSpan(ctx, &apiClient, LocalServerUrl+RoomMsgNoticePath, req, &res)
}
