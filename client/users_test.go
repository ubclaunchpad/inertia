package client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/res"
)

func TestUserClient_AddUser(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check correct endpoint called
		assert.Equal(t, "/user/add", r.URL.Path)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("uwu"))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer).GetUserClient()
	assert.NoError(t, d.AddUser(context.Background(), "", "", false))
}

func TestUserClient_RemoveUser(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check correct endpoint called
		assert.Equal(t, "/user/remove", r.URL.Path)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("uwu"))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer).GetUserClient()
	assert.NoError(t, d.RemoveUser(context.Background(), "yaoharry"))
}

func TestUserClient_ResetUser(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check correct endpoint called
		assert.Equal(t, "/user/reset", r.URL.Path)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("uwu"))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer).GetUserClient()
	assert.NoError(t, d.ResetUsers(context.Background()))
}

func TestUserClient_ListUsers(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "GET", r.Method)

		// Check correct endpoint called
		assert.Equal(t, "/user/list", r.URL.Path)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("users retrieved",
			"users", []string{"yaoharry"}))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer).GetUserClient()
	users, err := d.ListUsers(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"yaoharry"}, users)
}

func TestUserClient_Authenticate(t *testing.T) {
	username := "testguy"
	password := "SomeKindo23asdfpassword"

	t.Run("normal login", func(t *testing.T) {
		testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Check request method
			assert.Equal(t, http.MethodPost, r.Method)

			// Check correct endpoint called
			endpoint := r.URL.Path
			assert.Equal(t, "/user/login", endpoint)

			// Check auth
			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)
			assert.Equal(t, nil, err)
			var userReq api.UserRequest
			assert.Equal(t, nil, json.Unmarshal(body, &userReq))
			assert.Equal(t, userReq.Username, username)
			assert.Equal(t, userReq.Password, password)

			render.Render(w, r, res.MsgOK("session created",
				"token", "uwu"))
		}))
		defer testServer.Close()

		var d = newMockClient(t, testServer).GetUserClient()
		token, err := d.Authenticate(context.Background(), AuthenticateRequest{
			User:     username,
			Password: password,
			TOTP:     "",
		})
		assert.NoError(t, err)
		assert.Equal(t, "uwu", token)
	})

	t.Run("requires TOTP", func(t *testing.T) {
		testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			render.Render(w, r, res.Err("uwu", http.StatusPreconditionFailed))
		}))
		defer testServer.Close()

		var d = newMockClient(t, testServer).GetUserClient()
		token, err := d.Authenticate(context.Background(), AuthenticateRequest{
			User:     username,
			Password: password,
			TOTP:     "",
		})
		assert.Error(t, err)
		assert.Equal(t, "", token)
	})
}

func TestUserClient_EnableTotp(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check correct endpoint called
		endpoint := r.URL.Path
		assert.Equal(t, "/user/totp/enable", endpoint)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("TOTP successfully enabled",
			"totp", &api.TotpResponse{
				TotpSecret: "uwu",
			}))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer).GetUserClient()
	totp, err := d.EnableTotp(context.Background(), "", "")
	assert.NoError(t, err)
	assert.Equal(t, "uwu", totp.TotpSecret)
}

func TestUserClient_DisableTotp(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check request method
		assert.Equal(t, "POST", r.Method)

		// Check correct endpoint called
		assert.Equal(t, "/user/totp/disable", r.URL.Path)

		// Check auth
		assert.Equal(t, "Bearer "+fakeAuth, r.Header.Get("Authorization"))

		render.Render(w, r, res.MsgOK("uwu"))
	}))
	defer testServer.Close()

	var d = newMockClient(t, testServer).GetUserClient()
	assert.NoError(t, d.DisableTotp(context.Background()))
}
