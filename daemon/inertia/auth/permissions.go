package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ubclaunchpad/inertia/common"
)

// UserDatabasePath is the default location for storing users.
const UserDatabasePath = "/app/host/.inertia/users.db"

// PermissionsHandler handles users, permissions, and sessions on top
// of an http.ServeMux. It is used for Inertia Web.
type PermissionsHandler struct {
	domain      string
	users       *userManager
	sessions    *sessionManager
	mux         *http.ServeMux
	publicPaths []string
	adminPaths  []string
}

// NewPermissionsHandler returns a new handler for authenticating
// users and handling user administration
func NewPermissionsHandler(
	dbPath, domain, path string, timeout int,
	keyLookup ...func(*jwt.Token) (interface{}, error),
) (*PermissionsHandler, error) {
	// Set up user manager
	userManager, err := newUserManager(dbPath)
	if err != nil {
		return nil, err
	}

	// Set up session manager
	sessionManager := newSessionManager(domain, path, timeout)

	// Set up permissions handler
	mux := http.NewServeMux()
	handler := &PermissionsHandler{
		domain:     domain,
		users:      userManager,
		sessions:   sessionManager,
		mux:        mux,
		adminPaths: make([]string, 0),
	}

	// Set paths that don't require session authentication.
	handler.publicPaths = []string{
		"/login",
		"/adduser",
		"/removeuser",
		"/resetusers",
		"/listusers",
	}
	mux.HandleFunc("/login", handler.loginHandler)

	// The following endpoints are for user administration and must
	// be used from the CLI and delivered with the daemon token.
	lookup := GetAPIPrivateKey
	if len(keyLookup) > 0 {
		lookup = keyLookup[0]
	}
	mux.HandleFunc("/adduser", Authorized(handler.addUserHandler, lookup))
	mux.HandleFunc("/removeuser", Authorized(handler.removeUserHandler, lookup))
	mux.HandleFunc("/resetusers", Authorized(handler.resetUsersHandler, lookup))
	mux.HandleFunc("/listusers", Authorized(handler.listUsersHandler, lookup))

	return handler, nil
}

// Close releases resources held by the PermissionsHandler
func (h *PermissionsHandler) Close() error {
	h.sessions.Close()
	return h.users.Close()
}

// Implement the ServeHTTP method to make a permissionHandler a http.Handler
func (h *PermissionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Serve if path is public
	for _, prefix := range h.publicPaths {
		if strings.HasPrefix(path, prefix) {
			h.mux.ServeHTTP(w, r)
			return
		}
	}

	// Check if session is valid
	s, err := h.sessions.GetSession(w, r)
	if err != nil {
		if err == errSessionNotFound || err == errCookieNotFound {
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
				http.Error(w, "Admin privileges required", http.StatusForbidden)
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
	correct, err := h.users.IsCorrectCredentials(userReq.Username, userReq.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !correct {
		http.Error(w, "Login failed", http.StatusForbidden)
		return
	}
	err = h.sessions.BeginSession(userReq.Username, w, r)
	if err != nil {
		http.Error(w, "Login failed: "+err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] User %s logged in\n", http.StatusOK, userReq.Username)
}
