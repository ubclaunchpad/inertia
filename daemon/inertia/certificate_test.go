package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCertificate(t *testing.T) {
	dir := "./testcert/"
	os.Mkdir(dir, os.ModePerm)
	defer os.RemoveAll(dir)

	err := generateCertificate(dir+"test.cert", dir+"test.key", "0.0.0.0")
	assert.Nil(t, err)
	_, err = os.Stat(dir + "test.cert")
	assert.Nil(t, err)
	_, err = os.Stat(dir + "test.key")
	assert.Nil(t, err)
}
