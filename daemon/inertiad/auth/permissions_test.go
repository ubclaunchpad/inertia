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
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/api"
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
	user := &api.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
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
	req, err = http.NewRequest("GET", ts.URL+"/user/validate", nil)
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
	req, err = http.NewRequest("GET", ts.URL+"/user/validate", nil)
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
	user := &api.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
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
	user := &api.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
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
	user := &api.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
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
	body, err := json.Marshal(&api.UserRequest{
		Username: "jimmyneutron",
		Password: "asfasdlfjk",
		Admin:    false,
	})
	assert.Nil(t, err)
	payload := bytes.NewReader(body)
	req, err := http.NewRequest("POST", ts.URL+"/user/add", payload)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerTokenString)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Remove a user
	body, err = json.Marshal(&api.UserRequest{
		Username: "jimmyneutron",
	})
	assert.Nil(t, err)
	payload = bytes.NewReader(body)
	req, err = http.NewRequest("POST", ts.URL+"/user/remove", payload)
	assert.Nil(t, err)
	req.Header.Set("Authorization", bearerTokenString)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// List users
	req, err = http.NewRequest("GET", ts.URL+"/user/list", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", bearerTokenString)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Reset all users
	req, err = http.NewRequest("POST", ts.URL+"/user/reset", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", bearerTokenString)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEnableDisableTotpEndpoints(t *testing.T) {
	dir := "./test_enabledisable_totp"
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
	authToken := fmt.Sprintf("Bearer %s", crypto.TestMasterToken)

	// Add a new user
	body, err := json.Marshal(&api.UserRequest{
		Username: "jimmyneutron",
		Password: "asfasdlfjk",
		Admin:    false,
	})
	assert.Nil(t, err)
	payload := bytes.NewReader(body)
	req, err := http.NewRequest("POST", ts.URL+"/user/add", payload)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authToken)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Enable Totp
	payload = bytes.NewReader(body)
	req, err = http.NewRequest("POST", ts.URL+"/user/totp/enable", payload)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authToken)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Get Totp key from response
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	totpResp := &api.TotpResponse{}
	err = json.Unmarshal(respBytes, totpResp)
	assert.Nil(t, err)
	totpKey, err := totp.GenerateCode(totpResp.TotpSecret, time.Now())
	assert.Nil(t, err)

	// Log in with Totp
	body, err = json.Marshal(&api.UserRequest{
		Username: "jimmyneutron",
		Password: "asfasdlfjk",
		Totp:     totpKey,
	})
	payload = bytes.NewReader(body)
	req, err = http.NewRequest("POST", ts.URL+"/user/login", payload)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Get user JWT from response
	userTokenBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	authToken = fmt.Sprintf("Bearer %s", string(userTokenBytes))

	// Disable Totp
	req, err = http.NewRequest("POST", ts.URL+"/user/totp/disable", nil)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authToken)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPermissionsHandler_addUserHandler(t *testing.T) {
	type args struct {
		method string
		target string
		body   interface{}
	}
	type want struct {
		status int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"missing body", args{"POST", "/", nil}, want{http.StatusBadRequest}},
		{"bad credentials", args{"POST", "/", api.UserRequest{
			Username: "bobheadxi", Password: "bobheadxi",
		}}, want{http.StatusBadRequest}},
		{"ok credentials", args{"POST", "/", api.UserRequest{
			Username: "bobheadxi", Password: "bobdeadxi",
		}}, want{http.StatusCreated}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up permission handler
			var dir = "./test_addUserHandler"
			ph, err := getTestPermissionsHandler(dir)
			defer os.RemoveAll(dir)
			assert.Nil(t, err)
			defer ph.Close()

			// test handler
			var (
				b, _ = json.Marshal(tt.args.body)
				req  = httptest.NewRequest(tt.args.method, tt.args.target, bytes.NewReader(b))
				rec  = httptest.NewRecorder()
			)
			ph.addUserHandler(rec, req)

			// assert
			if rec.Code != tt.want.status {
				t.Errorf("expected status '%d', got '%d'", tt.want.status, rec.Code)
			}
		})
	}
}

func TestPermissionsHandler_loginHandler(t *testing.T) {
	type fields struct {
		user api.UserRequest
	}
	type args struct {
		method string
		target string
		body   interface{}
	}
	type want struct {
		status int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{"missing body", fields{}, args{"POST", "/", nil}, want{http.StatusBadRequest}},
		{"invalid user", fields{}, args{"POST", "/", api.UserRequest{
			Username: "bobhead", Password: "lunchpad",
		}}, want{http.StatusUnauthorized}},
		{"valid user, wrong creds", fields{api.UserRequest{
			Username: "bobhead", Password: "breakfastpad",
		}}, args{"POST", "/", api.UserRequest{
			Username: "bobhead", Password: "lunchpad",
		}}, want{http.StatusUnauthorized}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up permission handler
			var dir = "./test_loginHandler"
			ph, err := getTestPermissionsHandler(dir)
			defer os.RemoveAll(dir)
			assert.Nil(t, err)
			defer ph.Close()

			// test situation
			var testUser = tt.fields.user
			ph.users.AddUser(testUser.Username, testUser.Password, testUser.Admin)
			// todo: test totp situations?

			// test handler
			var (
				b, _ = json.Marshal(tt.args.body)
				req  = httptest.NewRequest(tt.args.method, tt.args.target, bytes.NewReader(b))
				rec  = httptest.NewRecorder()
			)
			ph.loginHandler(rec, req)

			// assert
			if rec.Code != tt.want.status {
				t.Logf("Received response: '%s'", rec.Body.String())
				t.Errorf("expected status '%d', got '%d'", tt.want.status, rec.Code)
			}
		})
	}
}
