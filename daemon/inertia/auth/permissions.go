package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ubclaunchpad/inertia/common"
	"github.com/xyproto/permissionbolt"
	"github.com/xyproto/pinterface"
)

// UserDatabasePath is the default location for storing users.
const UserDatabasePath = "/app/host/.inertia/users.db"

// PermissionsHandler handles users, permissions, and sessions on top
// of an http.ServeMux. It is used for Inertia Web.
type PermissionsHandler struct {
	// perm is a Permissions structure that can be used to deny requests
	// and acquire the UserState. By using `pinterface.IPermissions` instead
	// of `*permissionbolt.Permissions`, the code is compatible with not only
	// `permissionbolt`, but also other modules that uses other database
	// backends, like `permissions2` which uses Redis.
	perm pinterface.IPermissions

	// Mux is the HTTP multiplexer
	Mux *http.ServeMux
}

// Implement the ServeHTTP method to make a permissionHandler a http.Handler
func (ph *PermissionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if the user has the right admin/user rights
	if ph.perm.Rejected(w, r) {
		ph.perm.DenyFunction()(w, r)
		return
	}
	// Serve the requested page if permissions were granted
	ph.Mux.ServeHTTP(w, r)
}

// addUserHandler handles requests to add users
func (ph *PermissionsHandler) addUserHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve user details from request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var userReq common.UserRequest
	err = json.Unmarshal(body, &userReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add user (as admin if specified)
	userstate := ph.perm.UserState()
	userstate.AddUser(userReq.Username, userReq.Password, userReq.Email)
	userstate.MarkConfirmed(userReq.Username)
	if userReq.Admin {
		userstate.SetAdminStatus(userReq.Username)
	}
}

// removeUserHandler handles requests to add users
func (ph *PermissionsHandler) removeUserHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve user details from request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var userReq common.UserRequest
	err = json.Unmarshal(body, &userReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Remove user
	userstate := ph.perm.UserState()
	userstate.RemoveUser(userReq.Username)
}

// loginHandler handles requests to add users
func (ph *PermissionsHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve user details from request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var userReq common.UserRequest
	err = json.Unmarshal(body, &userReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log in user if password is correct
	userstate := ph.perm.UserState()
	correct := userstate.CorrectPassword(userReq.Username, userReq.Password)
	if correct {
		userstate.Login(w, userReq.Username)
	} else {
		http.Error(w, "Login failed", http.StatusForbidden)
	}
}

// logoutHandler handles requests to add users
func (ph *PermissionsHandler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve user details from request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var userReq common.UserRequest
	err = json.Unmarshal(body, &userReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log out user
	userstate := ph.perm.UserState()
	userstate.Logout(userReq.Username)
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	// @todo: direct to login page
	http.Error(w, "Permission denied!", http.StatusForbidden)
}

// NewPermissionsHandler returns a new handler for authenticating
// users and handling user administration
func NewPermissionsHandler(dbPath string) (*PermissionsHandler, error) {
	mux := http.NewServeMux()
	perm, err := permissionbolt.NewWithConf(dbPath)
	if err != nil {
		println("Could not open Bolt database")
		return nil, err
	}

	// Set permissions
	perm.Clear()
	perm.SetUserPath([]string{"/"})
	perm.SetPublicPath([]string{"/login"})

	// Set default handler for unauthenticated users
	perm.SetDenyFunction(redirectToLogin)

	// Set up webhandler
	ph := &PermissionsHandler{perm: perm, Mux: mux}

	// The following endpoints are for user administration and must
	// be used from the CLI and delivered with the daemon token.
	mux.HandleFunc("/adduser", Authorized(ph.addUserHandler, GetAPIPrivateKey))
	mux.HandleFunc("/removeuser", Authorized(ph.removeUserHandler, GetAPIPrivateKey))

	// The following endpoints are for the web app.
	mux.HandleFunc("/login", ph.loginHandler)
	mux.HandleFunc("/logout", ph.logoutHandler)

	return ph, nil
}
