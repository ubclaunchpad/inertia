package input

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
)

var (
	errInvalidInput  = errors.New("invalid input")
	errEmptyInput    = errors.New("empty input")
	errInvalidOption = errors.New("invalid option")

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

// PromptConfig offers prompt configuration
type PromptConfig struct {
	AllowEmpty   bool
	AllowInvalid bool
}

// PromptInteraction is a builder for interactions - use .PromptX followed by .GetX
type PromptInteraction struct {
	in   io.Reader
	conf PromptConfig
	resp string
	err  error
}

// NewPrompt instantiates a new prompt interaction on standard in
func NewPrompt(conf *PromptConfig) *PromptInteraction { return NewPromptOnInput(os.Stdin, conf) }

// NewPromptOnInput instantiates a new prompt on specified input
func NewPromptOnInput(in io.Reader, conf *PromptConfig) *PromptInteraction {
	if conf == nil {
		conf = &PromptConfig{}
	}
	return &PromptInteraction{in: in, conf: *conf}
}

func (p *PromptInteraction) parse() {
	var response string
	if _, err := fmt.Fscanln(p.in, &response); err != nil {
		if strings.Contains(err.Error(), "unexpected newline") {
			if !p.conf.AllowEmpty {
				p.err = errors.New("empty response not allowed")
			}
		} else {
			p.err = err
		}
	} else {
		p.resp = response
	}
}

// Prompt prints the given query and reads the response
func (p *PromptInteraction) Prompt(query ...interface{}) *PromptInteraction {
	out.Println(query...)
	p.parse()
	return p
}

// Promptf prints the given query and reads the response
func (p *PromptInteraction) Promptf(query string, args ...interface{}) *PromptInteraction {
	out.Printf(query+"\n", args...)
	p.parse()
	return p
}

// PromptFromList creates a choose-one-from-x prompt
func (p *PromptInteraction) PromptFromList(optionName string, options []string) *PromptInteraction {
	out.Printf("Available %ss:\n", optionName)
	for _, o := range options {
		out.Println("  > " + o)
	}
	out.Print(out.C("Please enter your desired %s: ", out.CY).With(optionName))
	p.parse()

	// check option is valid
	if p.err == nil && !p.conf.AllowInvalid {
		for _, o := range options {
			if o == p.resp {
				return p
			}
		}
		p.err = fmt.Errorf("invalid input %s: %w", p.resp, errInvalidOption)
	}
	return p
}

// GetBool retrieves a boolean response based on "y" or "yes"
func (p *PromptInteraction) GetBool() (bool, error) {
	yes := p.resp == "y"
	if !yes && !p.conf.AllowInvalid {
		if p.resp != "N" {
			return false, errInvalidInput
		}
	}
	return yes, p.err
}

// GetString retreives the raw string response from the prompt
func (p *PromptInteraction) GetString() (string, error) {
	return p.resp, p.err
}

// AddProjectWalkthrough is the command line walkthrough that asks for details
// about the project the user intends to deploy.
func AddProjectWalkthrough() (
	buildType cfg.BuildType, buildFilePath string, err error,
) {
	resp, err := NewPrompt(nil).
		PromptFromList("build type", []string{"docker-compose", "dockerfile"}).
		GetString()
	if err != nil {
		return "", "", errInvalidBuildType
	}
	buildType, err = cfg.AsBuildType(resp)
	if err != nil {
		return "", "", err
	}

	buildFilePath, err = NewPrompt(nil).
		Prompt(out.C("Please enter the path to your build configuration file:", out.CY)).
		GetString()
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
