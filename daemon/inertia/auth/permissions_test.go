package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/ubclaunchpad/inertia/common"

	"github.com/stretchr/testify/assert"
)

func getTestPermissionsHandler(dir string) (*PermissionsHandler, error) {
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return NewPermissionsHandler(path.Join(dir, "users.db"), "127.0.0.1", "/", 3000)
}

func TestServeHTTPPublicPath(t *testing.T) {
	dir := "./test"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachPublicHandler("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, err := http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServeHTTPWithUserReject(t *testing.T) {
	dir := "./test"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachUserRestrictedHandler("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, err := http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestServeHTTPWithUserLoginAndAccept(t *testing.T) {
	dir := "./test"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachUserRestrictedHandler("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", false)
	assert.Nil(t, err)

	// Login in as user, use cookiejar to catch cookie
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Check for cookies
	assert.True(t, len(loginResp.Cookies()) > 0)
	cookie := loginResp.Cookies()[0]
	assert.Equal(t, "ubclaunchpad-inertia", cookie.Name)

	// Attempt to access restricted endpoint with cookie
	req, err = http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServeHTTPDenyNonAdmin(t *testing.T) {
	dir := "./test"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachAdminRestrictedHandler("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", false)
	assert.Nil(t, err)

	// Login in as user, use cookiejar to catch cookie
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Check for cookies
	assert.True(t, len(loginResp.Cookies()) > 0)
	cookie := loginResp.Cookies()[0]
	assert.Equal(t, "ubclaunchpad-inertia", cookie.Name)

	// Attempt to access restricted endpoint with cookie
	req, err = http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestServeHTTPAllowAdmin(t *testing.T) {
	dir := "./test"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachAdminRestrictedHandler("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", true)
	assert.Nil(t, err)

	// Login in as user, use cookiejar to catch cookie
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Check for cookies
	assert.True(t, len(loginResp.Cookies()) > 0)
	cookie := loginResp.Cookies()[0]
	assert.Equal(t, "ubclaunchpad-inertia", cookie.Name)

	// Attempt to access restricted endpoint with cookie
	req, err = http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
