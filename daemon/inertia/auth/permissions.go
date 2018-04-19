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
	keyLookup   func(*jwt.Token) (interface{}, error)
	publicPaths []string
	adminPaths  []string
}

// NewPermissionsHandler returns a new handler for authenticating
// users and handling user administration. Param userlandPath is
// used to set cookie domain.
func NewPermissionsHandler(
	dbPath, domain, userlandPath string, timeout int,
	keyLookup ...func(*jwt.Token) (interface{}, error),
) (*PermissionsHandler, error) {
	// Set up user manager
	userManager, err := newUserManager(dbPath)
	if err != nil {
		return nil, err
	}

	// Set up session manager
	sessionManager := newSessionManager(domain, userlandPath, timeout)

	// Set up permissions handler
	mux := http.NewServeMux()
	handler := &PermissionsHandler{
		domain:   domain,
		users:    userManager,
		sessions: sessionManager,
		mux:      mux,
	}

	userPrefix := "/user/"
	userHandler := http.NewServeMux()

	// The following endpoints are for user administration and session
	// administration
	handler.publicPaths = []string{
		"/user/login",
		"/user/logout",
		"/user/validate",
	}
	userHandler.HandleFunc("/login", handler.loginHandler)
	userHandler.HandleFunc("/logout", handler.logoutHandler)
	userHandler.HandleFunc("/validate", handler.validateHandler)
	handler.adminPaths = []string{
		"/user/adduser",
		"/user/removeuser",
		"/user/resetusers",
		"/user/listusers",
	}
	handler.keyLookup = GetAPIPrivateKey
	if len(keyLookup) > 0 {
		handler.keyLookup = keyLookup[0]
	}
	userHandler.HandleFunc("/adduser", handler.addUserHandler)
	userHandler.HandleFunc("/removeuser", handler.removeUserHandler)
	userHandler.HandleFunc("/resetusers", handler.resetUsersHandler)
	userHandler.HandleFunc("/listusers", handler.listUsersHandler)
	mux.Handle(userPrefix, http.StripPrefix(userPrefix, userHandler))

	return handler, nil
}

// Close releases resources held by the PermissionsHandler
func (h *PermissionsHandler) Close() error {
	h.sessions.Close()
	return h.users.Close()
}

// Implement the ServeHTTP method to make a permissionHandler a http.Handler
func (h *PermissionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// http.StripPrefix removes the leading slash, but in the interest
	// of maintaining similar behaviour to stdlib handler functions,
	// we manually add a leading "/" here instead of having users not add
	// a leading "/" on the path if it dosn't already exist.
	path := r.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
		r.URL.Path = path
	}

	// Serve if path is public
	for _, prefix := range h.publicPaths {
		if strings.HasPrefix(path, prefix) {
			h.mux.ServeHTTP(w, r)
			return
		}
	}

	// Check if this path is for admins
	adminRestricted := false
	for _, prefix := range h.adminPaths {
		if strings.HasPrefix(path, prefix) {
			adminRestricted = true
		}
	}

	// Check token in header - if no tokens, check cookie
	bearerString := r.Header.Get("Authorization")
	splitToken := strings.Split(bearerString, "Bearer ")
	if len(splitToken) == 2 {
		tokenString := splitToken[1]

		// Parse takes the token string and a function for looking up the key.
		token, err := jwt.Parse(tokenString, h.keyLookup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// Verify the claims (none for now) and token.
		if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
			http.Error(w, tokenInvalidErrorMsg, http.StatusForbidden)
			return
		}

		// @todo: manage admin-restricted endpoints

		// Serve the requested endpoint to token holders
		h.mux.ServeHTTP(w, r)
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
	if adminRestricted {
		admin, err := h.users.IsAdmin(s.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if !admin {
			http.Error(w, "Admin privileges required", http.StatusForbidden)
		}
		return
	}

	// Serve the requested page if permissions were granted
	h.mux.ServeHTTP(w, r)
}

// AttachPublicHandler attaches given path and handler and makes it publicly available
func (h *PermissionsHandler) AttachPublicHandler(path string, handler http.HandlerFunc) {
	// Add path as exception to user restriction
	h.publicPaths = append(h.publicPaths, path)
	h.mux.Handle(path, handler)
}

// AttachUserRestrictedHandler attaches and restricts given path and handler to logged in users.
func (h *PermissionsHandler) AttachUserRestrictedHandler(path string, handler http.HandlerFunc) {
	// By default, all paths are user restricted
	h.mux.Handle(path, handler)
}

// AttachAdminRestrictedHandler attaches and restricts given path and handler to logged in admins.
func (h *PermissionsHandler) AttachAdminRestrictedHandler(path string, handler http.HandlerFunc) {
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
	// Handle CORS for local development
	if origin := r.Header.Get("Origin"); origin == "http://localhost:7900" {
		fmt.Println("setting cors")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			return
		}
	}

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

func (h *PermissionsHandler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	err := h.sessions.EndSession(w, r)
	if err != nil {
		http.Error(w, "Logout failed: "+err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] Session ended\n", http.StatusOK)
}

func (h *PermissionsHandler) validateHandler(w http.ResponseWriter, r *http.Request) {
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

	// Check if user exists
	err = h.users.HasUser(s.Username)
	if err != nil {
		if err == errUserNotFound {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Check if user is admin
	/*
		admin, err := h.users.IsAdmin(s.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	*/

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] Session ended\n", http.StatusOK)
}
