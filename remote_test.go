package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoteAddWalkthrough(t *testing.T) {
	mockCallback := func(name, address, sshPort, user, pemLoc, port string) error {
		assert.Equal(t, "pemfile", pemLoc)
		assert.Equal(t, "user", user)
		assert.Equal(t, "0.0.0.0", address)
		return nil
	}
	in, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	defer in.Close()

	fmt.Fprintln(in, "pemfile")
	fmt.Fprintln(in, "user")
	fmt.Fprintln(in, "0.0.0.0")

	_, err = in.Seek(0, os.SEEK_SET)
	assert.Nil(t, err)

	err = addRemoteWalkthrough(in, "inertia-rocks", "8080", "22", mockCallback)
	assert.Nil(t, err)
}

func TestRemoteAddWalkthroughFailure(t *testing.T) {
	mockCallback := func(name, address, sshPort, user, pemLoc, port string) error {
		return nil
	}
	in, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	defer in.Close()

	fmt.Fprintln(in, "pemfile")
	fmt.Fprintln(in, "")

	_, err = in.Seek(0, io.SeekStart)
	assert.Nil(t, err)

	err = addRemoteWalkthrough(in, "inertia-rocks", "8080", "22", mockCallback)
	assert.Equal(t, errInvalidUser, err)

	in.WriteAt([]byte("pemfile\nuser\n\n"), 0)
	_, err = in.Seek(0, io.SeekStart)
	assert.Nil(t, err)

	err = addRemoteWalkthrough(in, "inertia-rocks", "8080", "22", mockCallback)
	assert.Equal(t, errInvalidAddress, err)
}
