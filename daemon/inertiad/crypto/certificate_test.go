package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCertificateRSA(t *testing.T) {
	dir := "./testcert/"
	os.Mkdir(dir, os.ModePerm)
	defer os.RemoveAll(dir)

	err := GenerateCertificate(dir+"test.cert", dir+"test.key", "127.0.0.1:8081", "RSA")
	assert.NoError(t, err)
	_, err = os.Stat(dir + "test.cert")
	assert.NoError(t, err)
	_, err = os.Stat(dir + "test.key")
	assert.NoError(t, err)
}
