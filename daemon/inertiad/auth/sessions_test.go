package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/daemon/inertiad/crypto"
)

func Test_sessionManager_EndAllSessions(t *testing.T) {
	// make two sessions because map is pointer
	sessions := map[string]*crypto.TokenClaims{
		"1234": {User: "bob"},
	}
	manager := &sessionManager{internal: map[string]*crypto.TokenClaims{
		"1234": {User: "bob"},
	}}
	manager.EndAllSessions()
	assert.False(t, assert.ObjectsAreEqualValues(sessions, manager.internal))
}

func Test_sessionManager_EndAllUserSessions(t *testing.T) {
	// make two sessions because map is pointer
	sessions := map[string]*crypto.TokenClaims{
		"1234": {User: "bob"},
	}
	manager := &sessionManager{internal: map[string]*crypto.TokenClaims{
		"1234": {User: "bob"},
	}}
	manager.EndAllUserSessions("bob")
	assert.False(t, assert.ObjectsAreEqualValues(sessions, manager.internal))
}
