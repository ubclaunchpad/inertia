package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ubclaunchpad/inertia/common"
)

// UserDatabasePath is the default location for storing users.
const UserDatabasePath = "/app/host/.inertia/users.db"

// PermissionsHandler handles users, permissions, and sessions on top
// of an http.ServeMux. It is used for Inertia Web.
type PermissionsHandler struct {
	users       *userManager
	mux         *http.ServeMux
	publicPaths []string
	adminPaths  []string
}

// NewPermissionsHandler returns a new handler for authenticating
// users and handling user administration
func NewPermissionsHandler(dbPath string) (*PermissionsHandler, error) {
	// Set up user manager
	userManager, err := newUserManager(dbPath, 120)
	if err != nil {
		return nil, err
	}

	// Set up permissions handler
	mux := http.NewServeMux()
	handler := &PermissionsHandler{
		users:      userManager,
		mux:        mux,
		adminPaths: make([]string, 0),
	}

	// Set paths that don't require authentication.
	handler.publicPaths = []string{"/login", "/adduser", "/removeuser"}
	mux.HandleFunc("/login", handler.loginHandler)

	// The following endpoints are for user administration and must
	// be used from the CLI and delivered with the daemon token.
	mux.HandleFunc("/adduser", Authorized(handler.addUserHandler, GetAPIPrivateKey))
	mux.HandleFunc("/removeuser", Authorized(handler.removeUserHandler, GetAPIPrivateKey))

	return handler, nil
}

// Close releases resources held by the PermissionsHandler
func (h *PermissionsHandler) Close() error {
	return h.users.Close()
}

// Implement the ServeHTTP method to make a permissionHandler a http.Handler
func (h *PermissionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Serve if path is public
	for _, prefix := range h.publicPaths {
		if strings.HasPrefix(path, prefix) {
			h.mux.ServeHTTP(w, r)
		}
	}

	// Check if session is valid
	s, err := h.users.GetSession(w, r)
	if err != nil {
		if err == errSessionNotFound {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Check if user is valid
	err = h.users.HasUser(s.Username)
	if err != nil {
		if err == errUserNotFound {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Check if user has sufficient permissions for path
	for _, prefix := range h.adminPaths {
		if strings.HasPrefix(path, prefix) {
			admin, err := h.users.IsAdmin(s.Username)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if !admin {
				http.Error(w, err.Error(), http.StatusForbidden)
			}
			return
		}
	}

	// Serve the requested page if permissions were granted
	h.mux.ServeHTTP(w, r)
}

// AttachPublicHandler attaches given path and handler and makes it publicly available
func (h *PermissionsHandler) AttachPublicHandler(path string, handler http.Handler) {
	// Add path as exception to user restriction
	h.publicPaths = append(h.publicPaths, path)
	h.mux.Handle(path, handler)
}

// AttachUserRestrictedHandler attaches and restricts given path and handler to logged in users.
func (h *PermissionsHandler) AttachUserRestrictedHandler(path string, handler http.Handler) {
	// By default, all paths are user restricted
	h.mux.Handle(path, handler)
}

// AttachAdminRestrictedHandler attaches and restricts given path and handler to logged in admins.
func (h *PermissionsHandler) AttachAdminRestrictedHandler(path string, handler http.Handler) {
	// Add path as one that requires elevated permissions
	h.adminPaths = append(h.publicPaths, path)
	h.mux.Handle(path, handler)
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

	// Remove user
	err = h.users.RemoveUser(userReq.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] User %s removed\n", http.StatusOK, userReq.Username)
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
	correct, err := h.users.IsCorrectCredentials(userReq.Username, userReq.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !correct {
		http.Error(w, "Login failed", http.StatusForbidden)
		return
	}
	h.users.SessionBegin(userReq.Username, w, r)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] User %s logged in\n", http.StatusOK, userReq.Username)
}
