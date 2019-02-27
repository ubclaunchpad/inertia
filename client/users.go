package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ubclaunchpad/inertia/api"
)

var (
	// ErrNeedTotp is used to indicate that a 2FA-enabled user has not provided a TOTP
	ErrNeedTotp = errors.New("TOTP is needed for user")
)

// UserClient is used to access Inertia's /user APIs
type UserClient struct {
	c *Client
}

// NewUserClient instantiates a new client for user management functions
func NewUserClient(c *Client) *UserClient { return &UserClient{c} }

// AuthenticateRequest denotes options for authenticating with the Inertia daemon
type AuthenticateRequest struct {
	User     string
	Password string
	TOTP     string
}

// Authenticate gets an access token for the user with the given credentials. Use ""
// for totp if none is required.
func (u *UserClient) Authenticate(ctx context.Context, req AuthenticateRequest) (token string, err error) {
	resp, err := u.c.post(ctx, "/user/login", &api.UserRequest{
		Username: req.User,
		Password: req.Password,
		Totp:     req.TOTP,
	})
	if err != nil {
		return "", fmt.Errorf("failed to make request: %s", err.Error())
	}
	if resp.StatusCode == http.StatusExpectationFailed {
		return "", ErrNeedTotp
	}

	base, err := u.c.unmarshal(resp.Body, api.KV{Key: "token", Value: &token})
	resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed to read response: %s", err.Error())
	}

	return token, base.Error()
}

// AddUser adds an authorized user for access to Inertia Web
func (u *UserClient) AddUser(ctx context.Context, username, password string, admin bool) error {
	resp, err := u.c.post(ctx, "/user/add", &api.UserRequest{
		Username: username,
		Password: password,
		Admin:    admin,
	})
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}

	base, err := u.c.unmarshal(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}

	return base.Error()
}

// RemoveUser prevents a user from accessing Inertia Web
func (u *UserClient) RemoveUser(ctx context.Context, username string) error {
	resp, err := u.c.post(ctx, "/user/remove", &api.UserRequest{Username: username})
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}

	base, err := u.c.unmarshal(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}

	return base.Error()
}

// ResetUsers resets all users on the remote.
func (u *UserClient) ResetUsers(ctx context.Context) error {
	resp, err := u.c.post(ctx, "/user/reset", nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}

	base, err := u.c.unmarshal(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}

	return base.Error()
}

// ListUsers lists all users on the remote.
func (u *UserClient) ListUsers(ctx context.Context) ([]string, error) {
	resp, err := u.c.get(ctx, "/user/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %s", err.Error())
	}

	var users = make([]string, 0)
	base, err := u.c.unmarshal(resp.Body, api.KV{Key: "users", Value: &users})
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %s", err.Error())
	}

	return users, base.Error()
}

// EnableTotp enables Totp for a given user
func (u *UserClient) EnableTotp(ctx context.Context, username, password string) (*api.TotpResponse, error) {
	resp, err := u.c.post(ctx, "/user/totp/enable", &api.UserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %s", err.Error())
	}

	var totp api.TotpResponse
	base, err := u.c.unmarshal(resp.Body, api.KV{Key: "totp", Value: &totp})
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %s", err.Error())
	}

	return &totp, base.Error()
}

// DisableTotp disables Totp for a given user
func (u *UserClient) DisableTotp(ctx context.Context) error {
	resp, err := u.c.post(ctx, "/user/totp/disable", nil)
	if err != nil {
		return fmt.Errorf("failed to make request: %s", err.Error())
	}

	base, err := u.c.unmarshal(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err.Error())
	}

	return base.Error()
}
