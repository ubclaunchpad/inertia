package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

func getTestPermissionsHandler(dir string) (*PermissionsHandler, error) {
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return NewPermissionsHandler(
		path.Join(dir, "users.db"),
		"127.0.0.1", 3000,
		crypto.GetFakeAPIKey,
	)
}

func TestServeHTTPPublicPath(t *testing.T) {
	dir := "./test_perm_publicpath"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachPublicHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	dir := "./test_perm_reject"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachUserRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Without token
	req, err := http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	// With malformed token
	req.Header.Set("Authorization", "Bearer badtoken")
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestServeHTTPWithUserLoginAndLogout(t *testing.T) {
	dir := "./test_perm_loginlogout"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachUserRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", false)
	assert.Nil(t, err)

	// Login in as user
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Get token
	tokenBytes, err := ioutil.ReadAll(loginResp.Body)
	assert.Nil(t, err)
	token := string(tokenBytes)

	// Attempt to validate
	req, err = http.NewRequest("POST", ts.URL+"/user/validate", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Log out
	req, err = http.NewRequest("POST", ts.URL+"/user/logout", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	logoutResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer logoutResp.Body.Close()
	assert.Equal(t, http.StatusOK, logoutResp.StatusCode)

	// Attempt to validate again
	req, err = http.NewRequest("POST", ts.URL+"/user/validate", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestServeHTTPWithUserLoginAndAccept(t *testing.T) {
	dir := "./test_perm_loginaccept"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachUserRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", false)
	assert.Nil(t, err)

	// Login in as user
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Get token
	tokenBytes, err := ioutil.ReadAll(loginResp.Body)
	assert.Nil(t, err)
	token := string(tokenBytes)

	// Attempt to access restricted endpoint with cookie
	req, err = http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServeHTTPDenyNonAdmin(t *testing.T) {
	dir := "./test_perm_denynonadmin"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachAdminRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", false)
	assert.Nil(t, err)

	// Login in as user
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Get token
	tokenBytes, err := ioutil.ReadAll(loginResp.Body)
	assert.Nil(t, err)
	token := string(tokenBytes)

	// Attempt to access restricted endpoint with cookie
	req, err = http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestServeHTTPAllowAdmin(t *testing.T) {
	dir := "./test_perm_disallowadmin"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachAdminRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", true)
	assert.Nil(t, err)

	// Login in as user
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Get token
	tokenBytes, err := ioutil.ReadAll(loginResp.Body)
	assert.Nil(t, err)
	token := string(tokenBytes)

	// Attempt to access restricted endpoint with cookie
	req, err = http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUserControlHandlers(t *testing.T) {
	dir := "./test_perm_usercontrol"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph

	// Test handler uses the getFakeAPIToken keylookup, which will match with
	// the testToken
	bearerTokenString := fmt.Sprintf("Bearer %s", crypto.TestMasterToken)

	// Add a new user
	body, err := json.Marshal(&common.UserRequest{
		Username: "jimmyneutron",
		Password: "asfasdlfjk",
		Admin:    false,
	})
	assert.Nil(t, err)
	payload := bytes.NewReader(body)
	req, err := http.NewRequest("POST", ts.URL+"/user/adduser", payload)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerTokenString)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Remove a user
	body, err = json.Marshal(&common.UserRequest{
		Username: "jimmyneutron",
	})
	assert.Nil(t, err)
	payload = bytes.NewReader(body)
	req, err = http.NewRequest("POST", ts.URL+"/user/removeuser", payload)
	assert.Nil(t, err)
	req.Header.Set("Authorization", bearerTokenString)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// List users
	req, err = http.NewRequest("POST", ts.URL+"/user/listusers", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", bearerTokenString)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Reset all users
	req, err = http.NewRequest("POST", ts.URL+"/user/resetusers", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", bearerTokenString)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
