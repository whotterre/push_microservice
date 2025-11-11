package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/whotterre/push_microservice/internal/config"
)

type OneSignalClient struct {
	cfg *config.Config
}

func NewOneSignalClient(cfg *config.Config) *OneSignalClient {
	return &OneSignalClient{
		cfg: cfg,
	}
}

type OneSignalNotification struct {
	AppID              string                 `json:"app_id,omitempty"`
	IncludePlayerIDs   []string               `json:"include_player_ids,omitempty"`
	IncludeExternalIDs []string               `json:"include_external_user_ids,omitempty"` // Your own user IDs
	IncludeSegments    []string               `json:"included_segments,omitempty"`
	Contents           map[string]string      `json:"contents"`
	Headings           map[string]string      `json:"headings,omitempty"`
	Data               map[string]interface{} `json:"data,omitempty"`
}

type OneSignalResponse struct {
	ID         string      `json:"id"`
	Recipients int         `json:"recipients"`
	Errors     interface{} `json:"errors,omitempty"` // Can be []string or object
}

// GetErrors returns errors as a slice of strings regardless of the format
func (r *OneSignalResponse) GetErrors() []string {
	if r.Errors == nil {
		return []string{}
	}

	// If it's already a string slice
	if errSlice, ok := r.Errors.([]interface{}); ok {
		result := make([]string, 0, len(errSlice))
		for _, e := range errSlice {
			if str, ok := e.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}

	// If it's an object/map
	if errMap, ok := r.Errors.(map[string]interface{}); ok {
		result := make([]string, 0)
		for key, val := range errMap {
			result = append(result, fmt.Sprintf("%s: %v", key, val))
		}
		return result
	}

	return []string{}
}

func (c *OneSignalClient) SendPushNotification(notification *OneSignalNotification) (*OneSignalResponse, error) {
	apiUrl := "https://api.onesignal.com/notifications?c=push"

	payload, err := json.Marshal(notification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal notification: %w", err)
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Basic "+c.cfg.OneSignalKey)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("onesignal API error (status %d): %s", res.StatusCode, string(body))
	}

	var oneSignalRes OneSignalResponse
	if err := json.Unmarshal(body, &oneSignalRes); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &oneSignalRes, nil
}

func (c *OneSignalClient) SendToUsers(userIDs []string, title, message string, data map[string]interface{}) (*OneSignalResponse, error) {
	notification := &OneSignalNotification{
		AppID:            c.cfg.OneSignalAppID,
		IncludePlayerIDs: userIDs,
		Contents: map[string]string{
			"en": message,
		},
		Headings: map[string]string{
			"en": title,
		},
		Data: data,
	}

	return c.SendPushNotification(notification)
}

// SendToExternalUserIDs sends to users using your own user identifiers (like email or UUID)
// This requires you to set external_user_id on the OneSignal SDK in your app
func (c *OneSignalClient) SendToExternalUserIDs(externalUserIDs []string, title, message string, data map[string]interface{}) (*OneSignalResponse, error) {
	notification := &OneSignalNotification{
		AppID:              c.cfg.OneSignalAppID,
		IncludeExternalIDs: externalUserIDs,
		Contents: map[string]string{
			"en": message,
		},
		Headings: map[string]string{
			"en": title,
		},
		Data: data,
	}

	return c.SendPushNotification(notification)
}

func (c *OneSignalClient) SendToSegment(segment, title, message string, data map[string]interface{}) (*OneSignalResponse, error) {
	notification := &OneSignalNotification{
		AppID:           c.cfg.OneSignalAppID,
		IncludeSegments: []string{segment},
		Contents: map[string]string{
			"en": message,
		},
		Headings: map[string]string{
			"en": title,
		},
		Data: data,
	}

	return c.SendPushNotification(notification)
}

// Player represents a OneSignal device/player
type Player struct {
	ID                string                 `json:"id"`
	DeviceType        int                    `json:"device_type"`
	DeviceOS          string                 `json:"device_os"`
	LastActive        int64                  `json:"last_active"`
	InvalidIdentifier bool                   `json:"invalid_identifier"`
	ExternalUserID    string                 `json:"external_user_id,omitempty"`
	Tags              map[string]interface{} `json:"tags,omitempty"`
}

// PlayersResponse is the response from the View devices API
type PlayersResponse struct {
	TotalCount int      `json:"total_count"`
	Offset     int      `json:"offset"`
	Limit      int      `json:"limit"`
	Players    []Player `json:"players"`
}

// GetPlayers fetches players/devices subscribed to the app
func (c *OneSignalClient) GetPlayers(limit, offset int) (*PlayersResponse, error) {
	if limit <= 0 || limit > 300 {
		limit = 300 // OneSignal max
	}

	// OneSignal REST API endpoint for fetching subscriptions
	apiUrl := fmt.Sprintf("https://onesignal.com/api/v1/players?app_id=%s&limit=%d&offset=%d",
		c.cfg.OneSignalAppID, limit, offset)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Basic " + c.cfg.OneSignalKey)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("onesignal API error (status %d): %s", res.StatusCode, string(body))
	}

	var playersRes PlayersResponse
	if err := json.Unmarshal(body, &playersRes); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &playersRes, nil
}
