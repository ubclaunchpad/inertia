package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"github.com/ubclaunchpad/inertia/api"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

// ctxKey represents keys used in request contexts
type ctxKey int

const (
	errMalformedHeaderMsg        = "malformed authorization error"
	ctxUsername           ctxKey = iota
)

// PermissionsHandler handles users, permissions, and sessions on top
// of an http.ServeMux. It is used for Inertia Web.
type PermissionsHandler struct {
	domain     string
	users      *userManager
	sessions   *sessionManager
	mux        *chi.Mux
	userPaths  []string
	adminPaths []string
}

// NewPermissionsHandler returns a new handler for authenticating users and
// handling user administration. It also serves as the primary server for the
// Inertia daemon.
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

	// Set up handler
	var h = &PermissionsHandler{
		domain:   hostDomain,
		users:    userManager,
		sessions: sessionManager,
		mux:      chi.NewMux(),

		// paths restricted to users
		userPaths: []string{
			"/user/validate",
			"/user/totp/enable",
			"/user/totp/disable"},

		// paths restricted to administrators
		adminPaths: []string{
			"/user/add",
			"/user/remove",
			"/user/reset",
			"/user/list"},
	}

	// Register useful middleware
	h.mux.Use(
		cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}).Handler,
		middleware.RequestID,
		middleware.RealIP,
		middleware.Recoverer)

	// Register all user-related routes that managed by the permissions handler
	h.mux.Route("/user", func(r chi.Router) {
		r.Post("/login", h.loginHandler)
		r.Post("/logout", h.logoutHandler)

		// user-only paths
		r.Get("/validate", h.validateHandler)
		r.Route("/totp", func(r chi.Router) {
			r.Post("/enable", h.enableTotpHandler)
			r.Post("/disable", h.disableTotpHandler)
		})

		// admin-only paths
		r.Get("/list", h.listUsersHandler)
		r.Post("/add", h.addUserHandler)
		r.Post("/remove", h.removeUserHandler)
		r.Post("/reset", h.resetUsersHandler)
	})

	return h, nil
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
			http.Error(w, "admin privileges required", http.StatusForbidden)
			return
		}
	}

	// Attach username to request context so handlers can use it
	ctx := context.WithValue(r.Context(), ctxUsername, claims.User)

	// Serve the requested endpoint to token holders
	h.mux.ServeHTTP(w, r.WithContext(ctx))
}

// AttachPublicHandler attaches given path and handler and makes it publicly available
func (h *PermissionsHandler) AttachPublicHandler(path string, handler http.Handler) {
	h.mux.Handle(path, handler)
}

// AttachPublicHandlerFunc attaches given path and handler and makes it publicly available
func (h *PermissionsHandler) AttachPublicHandlerFunc(
	path string,
	handler http.HandlerFunc,
	methods ...string,
) {
	h.register(path, handler, methods)
}

// AttachUserRestrictedHandlerFunc attaches and restricts given path and handler to logged in users.
func (h *PermissionsHandler) AttachUserRestrictedHandlerFunc(
	path string,
	handler http.HandlerFunc,
	methods ...string,
) {
	h.userPaths = append(h.userPaths, path)
	h.register(path, handler, methods)
}

// AttachAdminRestrictedHandlerFunc attaches and restricts given path and handler to logged in admins.
func (h *PermissionsHandler) AttachAdminRestrictedHandlerFunc(
	path string,
	handler http.HandlerFunc,
	methods ...string,
) {
	h.adminPaths = append(h.adminPaths, path)
	h.register(path, handler, methods)
}

func (h *PermissionsHandler) register(path string, handler http.HandlerFunc, methods []string) {
	if len(methods) == 0 {
		h.mux.HandleFunc(path, handler)
	} else {
		for _, m := range methods {
			switch m {
			case http.MethodGet:
				h.mux.Get(path, handler)
			case http.MethodPost:
				h.mux.Post(path, handler)
			}
		}
	}
}

func (h *PermissionsHandler) addUserHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve user details from request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusLengthRequired)
		return
	}
	defer r.Body.Close()
	var userReq api.UserRequest
	err = json.Unmarshal(body, &userReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add user (as admin if specified)
	if err = h.users.AddUser(userReq.Username, userReq.Password, userReq.Admin); err != nil {
		if crypto.IsCredentialFormatError(err) {
			http.Error(w, fmt.Sprintf("Invalid credentials format: %s", err.Error()), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
	var userReq api.UserRequest
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

func (h *PermissionsHandler) enableTotpHandler(w http.ResponseWriter, r *http.Request) {
	userReq, err := readCredentials(r)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Check if password is correct (we do this first because we don't
	// want to reveal information about the user to the requester before they
	// are authenticated)
	_, correct, err := h.users.IsCorrectCredentials(
		userReq.Username, userReq.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !correct {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Make sure the user does not already have TOTP enabled
	totpEnabled, err := h.users.IsTotpEnabled(userReq.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if totpEnabled {
		http.Error(w, "Totp already enabled", http.StatusConflict)
		return
	}

	totpSecret, backupCodes, err := h.users.EnableTotp(userReq.Username)
	if err != nil {
		http.Error(w, "Failed to create TOTP keys", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(&api.TotpResponse{
		TotpSecret:  totpSecret,
		BackupCodes: backupCodes,
	})
	if err != nil {
		http.Error(w, "Failed to create TOTP keys", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (h *PermissionsHandler) disableTotpHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value(ctxUsername).(string)
	// Make sure that TOTP is actually enabled
	totpEnabled, err := h.users.IsTotpEnabled(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !totpEnabled {
		http.Error(w, "TOTP not enabled", http.StatusConflict)
		return
	}

	if err = h.users.DisableTotp(username); err != nil {
		http.Error(w, "Fail to disable TOTP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
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
	userReq, err := readCredentials(r)
	if err != nil {
		http.Error(w, "Login failed: Bad request", http.StatusBadRequest)
		return
	}

	// Check the password is correct
	props, correct, err := h.users.IsCorrectCredentials(
		userReq.Username, userReq.Password)
	if err == errMissingCredentials {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if err != nil && err != errUserNotFound {
		http.Error(w, "Login failed: "+err.Error(), http.StatusInternalServerError)
		return
	} else if !correct || err == errUserNotFound {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Make sure TOTP is valid if the user has TOTP enabled
	totpEnabled, err := h.users.IsTotpEnabled(userReq.Username)
	if err != nil {
		http.Error(w, "Unable to verify credentials", http.StatusInternalServerError)
		return
	} else if totpEnabled {
		if userReq.Totp == "" {
			http.Error(w, "Expected TOTP", http.StatusExpectationFailed)
			return
		}
		validTotp, err := h.users.IsValidTotp(userReq.Username, userReq.Totp)
		if err != nil {
			http.Error(w, "Unable to verify credentials", http.StatusInternalServerError)
			return
		} else if !validTotp {
			// Check if the user entered a backup code
			validBackup, err := h.users.IsValidBackupCode(userReq.Username, userReq.Totp)
			if err != nil {
				http.Error(w, "Unable to verify credentials", http.StatusInternalServerError)
				return
			} else if !validBackup {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}
		}
	}

	_, token, err := h.sessions.BeginSession(userReq.Username, props.Admin)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Write back
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
}

func (h *PermissionsHandler) logoutHandler(w http.ResponseWriter, r *http.Request) {
	err := h.sessions.EndSession(r)
	if err != nil {
		http.Error(w, "logout failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "[SUCCESS %d] Session ended\n", http.StatusOK)
}

func (h *PermissionsHandler) validateHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readCredentials(r *http.Request) (api.UserRequest, error) {
	userReq := api.UserRequest{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return userReq, err
	}
	defer r.Body.Close()
	err = json.Unmarshal(body, &userReq)
	return userReq, err
}
