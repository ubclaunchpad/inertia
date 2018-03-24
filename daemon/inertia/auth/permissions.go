package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ubclaunchpad/inertia/common"
)

// UserDatabasePath is the default location for storing users.
const UserDatabasePath = "/app/host/.inertia/users.db"

// PermissionsHandler handles users, permissions, and sessions on top
// of an http.ServeMux. It is used for Inertia Web.
type PermissionsHandler struct {
	users       *userManager
	mux         *http.ServeMux
	denyHandler http.Handler
	publicPaths []string
}

// NewPermissionsHandler returns a new handler for authenticating
// users and handling user administration
func NewPermissionsHandler(dbPath string, denyHandler http.HandlerFunc) (*PermissionsHandler, error) {
	// Set up user manager
	userManager, err := newUserManager(dbPath, 120)
	if err != nil {
		return nil, err
	}

	// Set up permissions handler
	mux := http.NewServeMux()
	handler := &PermissionsHandler{
		users:       userManager,
		mux:         mux,
		denyHandler: denyHandler,
	}

	// The following endpoints are for user administration and must
	// be used from the CLI and delivered with the daemon token.
	handler.publicPaths = []string{"/adduser", "/removeuser"}
	mux.HandleFunc("/adduser", Authorized(handler.addUserHandler, GetAPIPrivateKey))
	mux.HandleFunc("/removeuser", Authorized(handler.removeUserHandler, GetAPIPrivateKey))

	// The following endpoints require no prior authentication.
	mux.HandleFunc("/login", handler.loginHandler)

	return handler, nil
}

// Close releases resources held by the PermissionsHandler
func (handler *PermissionsHandler) Close() error {
	return handler.users.Close()
}

// Implement the ServeHTTP method to make a permissionHandler a http.Handler
func (handler *PermissionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if the user has the right admin/user rights

	// Serve the requested page if permissions were granted
	handler.mux.ServeHTTP(w, r)
}

// AttachUserRestrictedHandler attaches and restricts given path and handler to logged in users.
func (handler *PermissionsHandler) AttachUserRestrictedHandler(path string, h http.Handler) {
	handler.publicPaths = append(handler.publicPaths, path)
	// @todo
	handler.mux.Handle(path, h)
}

// AttachAdminRestrictedHandler attaches and restricts given path and handler to logged in admins.
func (handler *PermissionsHandler) AttachAdminRestrictedHandler(path string, h http.Handler) {
	// @todo
	handler.mux.Handle(path, h)
}

// User Administration Endpoint Handlers

func (handler *PermissionsHandler) addUserHandler(w http.ResponseWriter, r *http.Request) {
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
	err = handler.users.AddUser(userReq.Username, userReq.Password, userReq.Admin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "[SUCCESS %d] User %s added!\n", http.StatusCreated, userReq.Username)
}

func (handler *PermissionsHandler) removeUserHandler(w http.ResponseWriter, r *http.Request) {
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
	err = handler.users.RemoveUser(userReq.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] User %s removed\n", http.StatusOK, userReq.Username)
}

func (handler *PermissionsHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
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
	correct, err := handler.users.IsCorrectCredentials(userReq.Username, userReq.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !correct {
		http.Error(w, "Login failed", http.StatusForbidden)
	}
	err = handler.users.SessionBegin(userReq.Username, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] User %s logged in\n", http.StatusOK, userReq.Username)
}
