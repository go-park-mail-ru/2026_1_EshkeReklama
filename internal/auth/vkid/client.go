package vkid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrUnauthorized    = errors.New("vk id unauthorized")
	ErrUnexpectedReply = errors.New("vk id unexpected response")
)

type Identity struct {
	UserID int64
	Email  string
	Phone  string
}

type Client struct {
	baseURL     string
	clientID    int64
	redirectURI string
	httpClient  *http.Client
}

func New(clientID int64, redirectURI, domain string, timeout time.Duration) *Client {
	baseURL := strings.TrimSpace(domain)
	if baseURL == "" {
		baseURL = "https://id.vk.ru"
	} else if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		clientID:    clientID,
		redirectURI: redirectURI,
		httpClient:  &http.Client{Timeout: timeout},
	}
}

func (c *Client) ExchangeUser(ctx context.Context, code, deviceID, codeVerifier string) (*Identity, error) {
	token, err := c.exchangeCode(ctx, code, deviceID, codeVerifier)
	if err != nil {
		return nil, err
	}

	user, err := c.userInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	userID := token.UserID
	if user.User.UserID != "" {
		parsedID, parseErr := strconv.ParseInt(user.User.UserID, 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: parse user id: %v", ErrUnexpectedReply, parseErr)
		}
		userID = parsedID
	}

	if userID <= 0 {
		return nil, fmt.Errorf("%w: missing user id", ErrUnexpectedReply)
	}

	return &Identity{
		UserID: userID,
		Email:  strings.TrimSpace(strings.ToLower(user.User.Email)),
		Phone:  strings.TrimSpace(user.User.Phone),
	}, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	UserID      int64  `json:"user_id"`
}

type userInfoResponse struct {
	User struct {
		Email  string `json:"email"`
		Phone  string `json:"phone"`
		UserID string `json:"user_id"`
	} `json:"user"`
}

type errorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (c *Client) exchangeCode(ctx context.Context, code, deviceID, codeVerifier string) (*tokenResponse, error) {
	query := url.Values{
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {c.redirectURI},
		"client_id":     {strconv.FormatInt(c.clientID, 10)},
		"code_verifier": {codeVerifier},
		"device_id":     {deviceID},
	}

	var token tokenResponse
	if err := c.postForm(ctx, c.baseURL+"/oauth2/auth?"+query.Encode(), url.Values{"code": {code}}, &token); err != nil {
		return nil, err
	}

	if token.AccessToken == "" {
		return nil, fmt.Errorf("%w: empty access token", ErrUnexpectedReply)
	}

	return &token, nil
}

func (c *Client) userInfo(ctx context.Context, accessToken string) (*userInfoResponse, error) {
	query := url.Values{
		"client_id": {strconv.FormatInt(c.clientID, 10)},
	}

	var userInfo userInfoResponse
	if err := c.postForm(ctx, c.baseURL+"/oauth2/user_info?"+query.Encode(), url.Values{"access_token": {accessToken}}, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (c *Client) postForm(ctx context.Context, endpoint string, form url.Values, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vk id request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read vk id response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var apiErr errorResponse
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != "" {
			return fmt.Errorf("%w: %s", ErrUnauthorized, strings.TrimSpace(apiErr.ErrorDescription))
		}
		return fmt.Errorf("%w: status %d", ErrUnauthorized, resp.StatusCode)
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("%w: %v", ErrUnexpectedReply, err)
	}

	return nil
}
