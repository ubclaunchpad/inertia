package main

import (
	"net/http"

	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

// tokenHandler generates a new token
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	keyBytes, err := crypto.GetAPIPrivateKey(nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := crypto.GenerateMasterToken(keyBytes.([]byte))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
}
