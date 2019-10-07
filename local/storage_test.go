package local

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/inertia/cfg"
)

func TestWrite(t *testing.T) {
	type args struct {
		path    string
		data    interface{}
		writers []io.Writer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"nothing to write to", args{"", nil, nil}, true},
		{"ok: write to path", args{"./test-config.toml", &cfg.Remotes{
			Remotes: []*cfg.Remote{
				{
					Name: "dev",
					IP:   "0.0.0.0",
					SSH: &cfg.SSH{
						User: "bob",
					},
					Daemon: &cfg.Daemon{
						Port: "4043",
					},
					Profiles: map[string]string{
						"asdf": "asdf",
						"oipo": "oiup",
					}},
				{
					Name: "staging",
					IP:   "0.0.0.0",
					SSH: &cfg.SSH{
						User: "bob",
					},
					Daemon: &cfg.Daemon{
						Port: "4043",
					},
					Profiles: map[string]string{
						"fdsa":  "fdsaf",
						"wqrte": "erterh",
					}},
			},
		}, nil}, false},
		{"ok: write to path", args{"./test-config.2.toml", &cfg.Project{
			Name: "test",
			Profiles: []*cfg.Profile{
				{
					Name: "dev",
					Build: &cfg.Build{
						Type:          cfg.Dockerfile,
						BuildFilePath: "Dockerfile.dev",
					},
				},
				{
					Name: "staging",
					Build: &cfg.Build{
						Type:          cfg.Dockerfile,
						BuildFilePath: "Dockerfile.staging",
					},
				},
			},
		}, nil}, false},
		{"ok: write to writers", args{"", &cfg.Remotes{
			Remotes: make([]*cfg.Remote, 0),
		}, []io.Writer{os.Stdout}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.path != "" {
				defer os.RemoveAll(tt.args.path)
			}
			var err = Write(tt.args.path, tt.args.data, tt.args.writers...)
			assert.Equalf(t, (err != nil), tt.wantErr, "got '%v'", err)
		})
	}
}

func TestSaveKey(t *testing.T) {
	keyMaterial := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAw+14SQTAidfYPDizCYPv0gWq4+wFeInCrZGo4BFbMcP7xhH+
htmm0qx7ctYbCS0tQmCvCnt4W5jwhqH9v65/b1PWv1qQbXbJq0iyeSspgpaB8xq+
AkWoBkUOT8iaUzESDgJfEpC9q1s7dAUpmRDD0JMVzdsv1VQqpR22VWtnpcFtAkNk
3CIXiKFYJ5677dVSrc45dhO4R67LguSPxpXNRcg26/cFKWQO+y2StnYVEEUtvoWN
z2tGQu2hftJtjzzCFXckH8VTJ8EgX0+3Co5jXEbm1idFGFgcAP1WT3xuGh+wpCXM
LYVdF18VxGzZe0bxStZ/+bhsaYfFLyU8qL7RnQIDAQABAoIBAFELWLczjQU30I1Q
ktZ7yebhS0gOaFDtAydS2j0dUNCsFehfpx5Wx8fbaxEceYB5PIB5h85ZNncFM3Et
bs4sOzBsyKbMqnNtMIx2fMTcUsZexZAu3qwH7jHxvLLJ8vQ4lxRObM88KgjIqzYZ
sJRNOAJ95QYLBaVDtIQqXzLEQ9JvDnB5++i18eIF31UXbcjvhNn4M2Goku2EZ9T8
ny0KnRDh9W/Is6ndsBGkDEbXFVMCs6ubIeL7LdJ1W/QNK4HB3ZeRWHMR+lElp+o5
4BY+5bQN7RrTPQmzU0lD1UAIOuPNQeUiGQs4jsV4Oz21z6AWMgg/qAjn91LaWcCH
JnDv++ECgYEA/zJbzNhxF7Kk64U1//XWhtZ3EdlbiapLq26Z10emtED9FrPJxGCz
+fDR2BwWUEpZDY3TBMmjeQeO+VN++PYGMjFogZIKNIuOhu2Qs7u92nCLyeB1aeTm
h90/5II64qCy5KN2fvU6Q2cxNNrCs0Dchh1GYYCH7+IR5NkelTQWRuUCgYEAxItZ
8JYoxfegJmK3RpzYWrbuK2tP7msA9VNSbzMdgFpLG9I+bSJPuQfdOfnhfZG/YG40
MBpUH1X9Jn06Ie6YsbQTeEWUY4H5RKdNKSyyJYepw6C/ndRCuInGPaqQ6FSfccld
mwB3ziaIZVjSaaLGpDFaSgosW4a8hDBbe+4wvFkCgYEAhfGKmWPpSATt5uhORYBl
DvS2Hlo1X3ZQrTQp7wKejvGlZSsMddRD4qXxnjpvw8iiISkVXufus3GyK08Vz9ph
uiqQraFXVekB7/P1BUE/Ds4PsO/s8J3CGgGYrXllKtopyzO42D4iTIp3G0TO+ILM
vF/VNwvdTZ0cwz7qfGmQX7kCgYAJqOOpvGeSm0IGwPFLCihkBPudrK+IA0BPzmGN
z5BSn51zZ5jj2jza1jUcRVi8yC4EukXcW17pD1vayWrTAhwFF9mhHqJVZazvn91d
+bFjwNAqKjtgsW76DONuYnSuxoHzoLb2CEbbHe+0M3Jb+MEUjsxmOSvG789SG+JT
K/i/OQKBgQD3rq8dDSVYaLcSFwg9RfRKF+Ahtml86lm4FrfZlLEfwb6TaR/Unsh0
XF56ZdrKh0nbOW/125RSc8STCv5klDGnBCD56Qzbin9+W6j1TWyJFMdNeaxjWK+U
lq07qdr3cY+O1F4otlDitNuhLE88dtGJM5lEyumokiH1yXwhbBtZ4w==
-----END RSA PRIVATE KEY-----`
	cwd, _ := os.Getwd()
	testKeyPath := path.Join(cwd, "test_key_save")

	// Write
	err := SaveKey(keyMaterial, testKeyPath)
	assert.NoError(t, err)

	// Read
	bytes, err := ioutil.ReadFile(testKeyPath)
	assert.NoError(t, err)
	assert.Equal(t, keyMaterial, string(bytes))

	// Test config remove
	err = os.Remove(testKeyPath)
	assert.NoError(t, err)
}
