package input

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
)

var (
	errInvalidInput = errors.New("invalid input")
	errEmptyInput   = errors.New("empty input")

	errInvalidUser          = errors.New("invalid user")
	errInvalidAddress       = errors.New("invalid IP address")
	errInvalidBuildType     = errors.New("invalid build type")
	errInvalidBuildFilePath = errors.New("invalid buildfile path")
)

// CatchSigterm listens in the background for some kind of interrupt and calls
// the given cancelFunc as necessary
func CatchSigterm(cancelFunc func()) {
	var signals = make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		cancelFunc()
	}()
}

// Prompt prints the given query and reads the response
func Prompt(query ...interface{}) (string, error) {
	out.Println(query...)
	var response string
	if _, err := fmt.Fscanln(os.Stdin, &response); err != nil {
		if strings.Contains(err.Error(), "unexpected newline") {
			return "", nil
		}
		return "", err
	}
	return response, nil
}

// Promptf prints the given query and reads the response
func Promptf(query string, args ...interface{}) (string, error) {
	out.Printf(query+"\n", args...)
	var response string
	if _, err := fmt.Fscanln(os.Stdin, &response); err != nil {
		return "", err
	}
	return response, nil
}

// AddProjectWalkthrough is the command line walkthrough that asks for details
// about the project the user intends to deploy
func AddProjectWalkthrough() (
	buildType cfg.BuildType, buildFilePath string, err error,
) {
	out.Println(out.C("Please enter the path to your build configuration file:", out.CY))
	out.Println("  - docker-compose")
	out.Println("  - dockerfile")

	var response string
	if _, err = fmt.Fscanln(os.Stdin, &response); err != nil {
		return "", "", errInvalidBuildType
	}
	buildType, err = cfg.AsBuildType(response)
	if err != nil {
		return "", "", err
	}

	buildFilePath, err = Prompt(
		out.C("Please enter the path to your build configuration file:", out.CY).String(),
	)
	if err != nil || buildFilePath == "" {
		return "", "", errInvalidBuildFilePath
	}
	return
}

// EnterEC2CredentialsWalkthrough prints promts to stdout and reads input from
// given reader
func EnterEC2CredentialsWalkthrough() (id, key string, err error) {
	out.Print(`To get your credentials:
	1. Open the IAM console (https://console.aws.amazon.com/iam/home?#home).
	2. In the navigation pane of the console, choose Users. You may have to create a user.
	3. Choose your IAM user name (not the check box).
	4. Choose the Security credentials tab and then choose Create access key.
	5. To see the new access key, choose Show. Your credentials will look something like this:

		Access key ID: AKIAIOSFODNN7EXAMPLE
		Secret access key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
	`)

	var response string

	out.Print("\nKey ID:       ")
	_, err = fmt.Fscanln(os.Stdin, &response)
	if err != nil {
		return
	}
	id = response

	out.Print("\nAccess Key:   ")
	_, err = fmt.Fscanln(os.Stdin, &response)
	if err != nil {
		return
	}
	key = response
	return
}

// ChooseFromListWalkthrough prints given options and reads in a choice from
// the given reader
func ChooseFromListWalkthrough(optionName string, options []string) (string, error) {
	out.Printf("Available %ss:\n", optionName)
	for _, o := range options {
		out.Println("  > " + o)
	}
	out.Printf("Please enter your desired %s: ", optionName)

	var response string
	_, err := fmt.Fscanln(os.Stdin, &response)
	if err != nil {
		return "", errInvalidInput
	}

	return response, nil
}
