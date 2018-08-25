package crypto

import "testing"

var (
	testSignature = "sha1=126f2c800419c60137ce748d7672e77b65cf16d6"
	testPayload   = []byte(`{"yo":true}`)
	testKey       = []byte("0123456789abcdef")
)

func TestValidateSignature(t *testing.T) {
	type args struct {
		signature string
		payload   []byte
		secretKey []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ok", args{testSignature, testPayload, testKey}, false},
		{"missing sig", args{"", testPayload, testKey}, true},
		{"incorrect sig", args{testSignature, testPayload, []byte("ohno")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateSignature(tt.args.signature, tt.args.payload, tt.args.secretKey); (err != nil) != tt.wantErr {
				t.Errorf("ValidateSignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
