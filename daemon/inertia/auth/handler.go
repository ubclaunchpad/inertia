package auth

import (
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ubclaunchpad/inertia/common"
)

// Authorized is a function decorator for authorizing RESTful
// daemon requests. It wraps handler functions and ensures the
// request is authorized. Returns a function
func Authorized(handler http.HandlerFunc, keyLookup func(*jwt.Token) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Collect the token from the header.
		bearerString := r.Header.Get("Authorization")

		// Split out the actual token from the header.
		splitToken := strings.Split(bearerString, "Bearer ")
		if len(splitToken) < 2 {
			http.Error(w, malformedAuthStringErrorMsg, http.StatusForbidden)
			return
		}
		tokenString := splitToken[1]

		// Parse takes the token string and a function for looking up the key.
		token, err := jwt.Parse(tokenString, keyLookup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// Verify the claims (none for now) and token.
		if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
			http.Error(w, tokenInvalidErrorMsg, http.StatusForbidden)
			return
		}

		// We're authorized, run the handler.
		handler(w, r)
	}
}

// HealthCheckHandler returns a 200 if the daemon is happy.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, common.MsgDaemonOK)
}
