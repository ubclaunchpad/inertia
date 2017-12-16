# 👩‍🚀 Inertia

Inertia makes it easy to set up automated deployment for Dockerized
applications.

## Installation

We use [dep](https://github.com/golang/dep) for managing dependencies. Install
that first if you haven't already.

```
brew install dep
```

Install project dependencies.

```bash
dep ensure
```

## Bootstrapping

To bootstrap servers, often some bash scripting is involved,
but we'd like to avoid shipping bash scripts with our go binary.
So we use [go-bindata](https://github.com/jteeuwen/go-bindata) to
compile shell scripts into our go executables.

```bash
go get -u github.com/jteeuwen/go-bindata/...
```

If you make changes to the bootstrapping shell scripts in
`cmd/bootstrap/`, convert them to `Assets` by running.

```bash
go-bindata -o cmd/bootstrap.go cmd/bootstrap/...
```

Inspect the auto-generated file `cmd/bootstrap.go`. Change its
package from `main` to `cmd`. Then use your asset!

```go
shellScriptData, err := Asset("cmd/bootstrap/myshellscript.sh")

if err != nil {
  log.Fatal("No asset with that name")
}

// Optionally run shell script over SSH.
result, _ := remote.RunSSHCommand(string(shellScriptData))
```

## Motivation

At Launch Pad we are frequently changing hosting providers based on available
funding and sponsorship. Inertia is a project to develop an in-house continuous
deployment system to make deploying applications simple and painless, regardless
of the hosting provider.

## Design

Inertia will contain two major components:

* Deployment daemon
* Command line interface

The deployment daemon will run persistently in the background, receiving webhook
events from GitHub whenever new commits are pushed. The CLI will provide an
interface to adjust settings, add repositories, etc.

This design differs from other similar tools because Inertia runs on the same
server as the project it is deploying.

### Setup

A primary design goal of Inertia is to minimize setup time for new projects. The
current setup flow is:

* Install and run Inertia on a new server
* Inertia will generate:
  * A SSH public key
  * A webhook URL and secret
* Add the SSH key to your project's Deploy Keys on GitHub
* Create a webhook with the URL and secret on your project repository


### Testing

+ Build the test image.

```bash
docker build -f ./test/Dockerfile -t inertia-test .
```

+ Run the tests.

```bash
docker run inertia-test
```
