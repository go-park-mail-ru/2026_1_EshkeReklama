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
	UserID    int64
	Email     string
	Phone     string
	FirstName string
	LastName  string
}

type Client struct {
	baseURL    string
	clientID   int64
	httpClient *http.Client
}

func New(clientID int64, domain string, timeout time.Duration) *Client {
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
		baseURL:    strings.TrimRight(baseURL, "/"),
		clientID:   clientID,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) ResolveUser(ctx context.Context, accessToken string) (*Identity, error) {
	user, err := c.userInfo(ctx, strings.TrimSpace(accessToken))
	if err != nil {
		return nil, err
	}

	userID, err := strconv.ParseInt(strings.TrimSpace(user.User.UserID), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: parse user id: %v", ErrUnexpectedReply, err)
	}
	if userID <= 0 {
		return nil, fmt.Errorf("%w: missing user id", ErrUnexpectedReply)
	}

	return &Identity{
		UserID:    userID,
		Email:     strings.TrimSpace(strings.ToLower(user.User.Email)),
		Phone:     strings.TrimSpace(user.User.Phone),
		FirstName: strings.TrimSpace(user.User.FirstName),
		LastName:  strings.TrimSpace(user.User.LastName),
	}, nil
}

type userInfoResponse struct {
	User struct {
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		UserID    string `json:"user_id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"user"`
}

type errorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
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
