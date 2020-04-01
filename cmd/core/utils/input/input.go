package input

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
)

var (
	// ErrEmptyInput is returned on empty imputs - toggle with AllowEmpty
	ErrEmptyInput = errors.New("empty input")
	// ErrInvalidInput is returned on disallowed inputs - toggle with AllowInvalid
	ErrInvalidInput = errors.New("invalid input")
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
		p.err = fmt.Errorf("illegal option '%s' chosen: %w", p.resp, ErrInvalidInput)
	}
	return p
}

// GetBool retrieves a boolean response based on "y" or "yes"
func (p *PromptInteraction) GetBool() (bool, error) {
	yes := p.resp == "y"
	if !yes && !p.conf.AllowInvalid {
		if p.resp != "N" && p.resp != "" {
			return false, fmt.Errorf("illegal input '%s' provided: %w", p.resp, ErrInvalidInput)
		}
	}
	return yes, p.err
}

// GetString retreives the raw string response from the prompt
func (p *PromptInteraction) GetString() (string, error) {
	return p.resp, p.err
}
