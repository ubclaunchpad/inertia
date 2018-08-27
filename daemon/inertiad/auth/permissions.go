package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

const (
	errMalformedHeaderMsg = "malformed authorization error"
)

// PermissionsHandler handles users, permissions, and sessions on top
// of an http.ServeMux. It is used for Inertia Web.
type PermissionsHandler struct {
	domain     string
	users      *userManager
	sessions   *sessionManager
	mux        *http.ServeMux
	userPaths  []string
	adminPaths []string
}

// NewPermissionsHandler returns a new handler for authenticating
// users and handling user administration. Param userlandPath is
// used to set cookie domain.
func NewPermissionsHandler(
	dbPath, hostDomain string, timeout int,
	keyLookup ...func(*jwt.Token) (interface{}, error),
) (*PermissionsHandler, error) {
	// Set up user manager
	userManager, err := newUserManager(dbPath)
	if err != nil {
		return nil, err
	}

	// Set up session manager
	lookup := crypto.GetAPIPrivateKey
	if len(keyLookup) > 0 {
		lookup = keyLookup[0]
	}
	sessionManager := newSessionManager(hostDomain, timeout, lookup)

	// Set up permissions handler
	mux := http.NewServeMux()
	handler := &PermissionsHandler{
		domain:   hostDomain,
		users:    userManager,
		sessions: sessionManager,
		mux:      mux,
	}

	// The following endpoints are for user administration and session administration
	userHandler := http.NewServeMux()
	userHandler.HandleFunc("/login", handler.loginHandler)
	userHandler.HandleFunc("/logout", handler.logoutHandler)
	handler.userPaths = []string{
		"/user/validate",
	}
	userHandler.HandleFunc("/validate", handler.validateHandler)
	handler.adminPaths = []string{
		"/user/adduser",
		"/user/removeuser",
		"/user/resetusers",
		"/user/listusers",
	}
	userHandler.HandleFunc("/adduser", handler.addUserHandler)
	userHandler.HandleFunc("/removeuser", handler.removeUserHandler)
	userHandler.HandleFunc("/resetusers", handler.resetUsersHandler)
	userHandler.HandleFunc("/listusers", handler.listUsersHandler)
	mux.Handle("/user/", http.StripPrefix("/user", userHandler))

	return handler, nil
}

// Close releases resources held by the PermissionsHandler
func (h *PermissionsHandler) Close() error {
	h.sessions.Close()
	return h.users.Close()
}

// nolint: gocyclo
func (h *PermissionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// http.StripPrefix removes the leading slash, but in the interest of
	// maintaining similar behaviour to stdlib handler functions, we manually
	// add a leading "/" here instead of having users not add a leading "/" on
	// the path if it dosn't already exist.
	path := r.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
		r.URL.Path = path
	}

	// Check if this path is restricted
	adminRestricted := false
	for _, prefix := range h.adminPaths {
		if strings.HasPrefix(path, prefix) {
			adminRestricted = true
		}
	}
	userRestricted := false
	for _, prefix := range h.userPaths {
		if strings.HasPrefix(path, prefix) {
			userRestricted = true
		}
	}

	// Serve directly if path is public
	if !userRestricted && !adminRestricted {
		h.mux.ServeHTTP(w, r)
		return
	}

	// Check if token is valid
	claims, err := h.sessions.GetSession(r)
	if err != nil {
		if err == errSessionNotFound {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}

	// Check if user has sufficient permissions for path
	if adminRestricted {
		admin, err := h.users.IsAdmin(claims.User)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !admin {
			http.Error(w, "Admin privileges required", http.StatusForbidden)
			return
		}
	}

	// Serve the requested endpoint to token holders
	h.mux.ServeHTTP(w, r)
}

// AttachPublicHandler attaches given path and handler and makes it publicly available
func (h *PermissionsHandler) AttachPublicHandler(path string, handler http.Handler) {
	h.mux.Handle(path, handler)
}

// AttachPublicHandlerFunc attaches given path and handler and makes it publicly available
func (h *PermissionsHandler) AttachPublicHandlerFunc(path string, handler http.HandlerFunc) {
	h.mux.HandleFunc(path, handler)
}

// AttachUserRestrictedHandlerFunc attaches and restricts given path and handler to logged in users.
func (h *PermissionsHandler) AttachUserRestrictedHandlerFunc(path string, handler http.HandlerFunc) {
	// Add path to user-restricted paths
	h.userPaths = append(h.userPaths, path)
	h.mux.HandleFunc(path, handler)
}

// AttachAdminRestrictedHandlerFunc attaches and restricts given path and handler to logged in admins.
func (h *PermissionsHandler) AttachAdminRestrictedHandlerFunc(path string, handler http.HandlerFunc) {
	// Add path as one that requires elevated permissions
	h.adminPaths = append(h.adminPaths, path)
	h.mux.HandleFunc(path, handler)
}

func (h *PermissionsHandler) addUserHandler(w http.ResponseWriter, r *http.Request) {
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
	err = h.users.AddUser(userReq.Username, userReq.Password, userReq.Admin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "[SUCCESS %d] User %s added!\n", http.StatusCreated, userReq.Username)
}

func (h *PermissionsHandler) removeUserHandler(w http.ResponseWriter, r *http.Request) {
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

	// Remove user credentials
	err = h.users.RemoveUser(userReq.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// End user sessions
	h.sessions.EndAllUserSessions(userReq.Username)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] User %s removed\n", http.StatusOK, userReq.Username)
}

func (h *PermissionsHandler) resetUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Delete all users
	err := h.users.Reset()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete all sessions
	h.sessions.EndAllSessions()

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] User and session databases reset\n", http.StatusOK)
}

func (h *PermissionsHandler) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	users := h.users.UserList()
	userList := ""
	for _, user := range users {
		userList += " - " + user + "\n"
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if len(users) != 0 {
		fmt.Fprintf(w, "[SUCCESS %d] Users: \n%s\n", http.StatusOK, userList)
	} else {
		fmt.Fprintf(w, "[SUCCESS %d] No users registered.", http.StatusOK)
	}
}

func (h *PermissionsHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
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
	props, correct, err := h.users.IsCorrectCredentials(userReq.Username, userReq.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !correct {
		http.Error(w, "Login failed", http.StatusForbidden)
		return
	}
	_, token, err := h.sessions.BeginSession(userReq.Username, props.Admin)
	if err != nil {
		http.Error(w, "Login failed: "+err.Error(), http.StatusForbidden)
		return
	}

	// Write back
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
}

func (h *PermissionsHandler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	err := h.sessions.EndSession(r)
	if err != nil {
		http.Error(w, "Logout failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] Session ended\n", http.StatusOK)
}

func (h *PermissionsHandler) validateHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
