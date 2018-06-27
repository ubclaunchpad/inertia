package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/common"
)

var (
	errInvalidInput = errors.New("invalid input")

	errInvalidUser          = errors.New("invalid user")
	errInvalidAddress       = errors.New("invalid IP address")
	errInvalidBuildType     = errors.New("invalid build type")
	errInvalidBuildFilePath = errors.New("invalid buildfile path")
)

// addRemoteWalkthough is the command line walkthrough that asks
// users for RemoteVPS details. It is up to the caller to save config.
func addRemoteWalkthrough(
	in io.Reader, config *cfg.Config,
	name, port, sshPort, currBranch string,
) error {
	homeEnvVar := os.Getenv("HOME")
	sshDir := filepath.Join(homeEnvVar, ".ssh")
	defaultSSHLoc := filepath.Join(sshDir, "id_rsa")

	var response string
	fmt.Println("Enter location of PEM file (leave blank to use '" + defaultSSHLoc + "'):")
	_, err := fmt.Fscanln(in, &response)
	if err != nil {
		response = defaultSSHLoc
	}
	pemLoc := response

	fmt.Println("Enter user:")
	n, err := fmt.Fscanln(in, &response)
	if err != nil || n == 0 {
		return errInvalidUser
	}
	user := response

	fmt.Println("Enter IP address of remote:")
	n, err = fmt.Fscanln(in, &response)
	if err != nil || n == 0 {
		return errInvalidAddress
	}
	address := response

	fmt.Println("Enter webhook secret (leave blank to generate one):")
	n, err = fmt.Fscanln(in, &response)
	if err != nil || n == 0 {
		response, err = common.GenerateRandomString()
		if err != nil {
			return err
		}
	}
	secret := response

	branch := currBranch
	fmt.Println("Enter project branch to deploy (leave blank for current branch):")
	n, err = fmt.Fscanln(in, &response)
	if err == nil && n != 0 {
		branch = response
	}

	fmt.Println("\nPort " + port + " will be used as the daemon port.")
	fmt.Println("Port " + sshPort + " will be used as the SSH port.")
	fmt.Println("Run 'inertia remote add' with the -p flag to set a custom Daemon port")
	fmt.Println("of the -ssh flag to set a custom SSH port.")

	config.AddRemote(&cfg.RemoteVPS{
		Name:    name,
		IP:      address,
		User:    user,
		PEM:     pemLoc,
		Branch:  branch,
		SSHPort: sshPort,
		Daemon: &cfg.DaemonConfig{
			Port:          port,
			WebHookSecret: secret,
		},
	})
	return nil
}

// addProjectWalkthrough is the command line walkthrough that asks for details
// about the project the user intends to deploy
func addProjectWalkthrough(in io.Reader) (buildType string, buildFilePath string, inputErr error) {
	println("Please enter the build type of your project - this could be one of:")
	println("  - docker-compose")
	println("  - dockerfile")
	println("  - herokuish")

	var response string
	_, err := fmt.Fscanln(in, &response)
	if err != nil {
		return "", "", errInvalidBuildType
	}
	buildType = response

	switch buildType {
	case "herokuish":
		return
	default:
		_, err := fmt.Fscanln(in, &response)
		if err != nil {
			return "", "", errInvalidBuildFilePath
		}
		buildFilePath = response
	}
	return
}

func enterEC2CredentialsWalkthrough(in io.Reader) (id, key string, err error) {
	print(`To get your credentials:
	1. Open the IAM console (https://console.aws.amazon.com/iam/home?#home).
	2. In the navigation pane of the console, choose Users. You may have to create a user.
	3. Choose your IAM user name (not the check box).
	4. Choose the Security credentials tab and then choose Create access key.
	5. To see the new access key, choose Show. Your credentials will look something like this:

		Access key ID: AKIAIOSFODNN7EXAMPLE
		Secret access key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
	`)

	var response string

	print("\nKey ID:       ")
	_, err = fmt.Fscanln(in, &response)
	if err != nil {
		return
	}
	id = response

	print("\nAccess Key:   ")
	_, err = fmt.Fscanln(in, &response)
	if err != nil {
		return
	}
	key = response
	return
}

func chooseFromListWalkthrough(in io.Reader, optionName string, options []string) (string, error) {
	fmt.Printf("Available %ss:\n", optionName)
	for _, o := range options {
		println("  > " + o)
	}
	fmt.Printf("Please enter your desired %s: ", optionName)

	var response string
	_, err := fmt.Fscanln(in, &response)
	if err != nil {
		return "", errInvalidInput
	}

	return response, nil
}
