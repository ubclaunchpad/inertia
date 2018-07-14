package crypto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateMasterToken(t *testing.T) {
	token, err := GenerateMasterToken(TestPrivateKey)
	assert.Nil(t, err)
	assert.Equal(t, TestMasterToken, token)

	otherToken, err := GenerateMasterToken([]byte("another_sekrit_key"))
	assert.Nil(t, err)
	assert.NotEqual(t, token, otherToken)

	// Verify validity
	readClaims, err := ValidateToken(token, GetFakeAPIKey)
	assert.Nil(t, err)
	assert.Nil(t, readClaims.Valid())
}

func TestTokenClaims_Valid(t *testing.T) {
	type fields struct {
		SessionID string
		User      string
		Admin     bool
		Expiry    time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// master key
		{"success", fields{"1234", "master", true, time.Time{}}, false},
		// expiry in future (+1)
		{"success", fields{"1234", "bob", true, time.Now().AddDate(0, 1, 0)}, false},
		// expiry in past (-1)
		{"fail", fields{"1234", "bob", true, time.Now().AddDate(0, -1, 0)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &TokenClaims{
				SessionID: tt.fields.SessionID,
				User:      tt.fields.User,
				Admin:     tt.fields.Admin,
				Expiry:    tt.fields.Expiry,
			}
			if err := claims.Valid(); (err != nil) != tt.wantErr {
				t.Errorf("TokenClaims.Valid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTokenClaims_GenerateToken(t *testing.T) {
	expires := time.Now().AddDate(0, 1, 0)
	claims := &TokenClaims{"1234", "robert", true, expires}
	token, err := claims.GenerateToken(TestPrivateKey)
	assert.Nil(t, err)

	// Try decoding token
	readClaims, err := ValidateToken(token, GetFakeAPIKey)
	assert.Nil(t, err)
	assert.Equal(t, claims.User, readClaims.User)
}
