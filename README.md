# 👩‍🚀 Inertia

Inertia makes it easy to set up automated deployment for Dockerized
applications.

## Installation

```bash
go get -u github.com/ubclauncpad/inertia
```

Alternatively, you can download Inertia from the [Releases](https://github.com/ubclaunchpad/inertia/releases) page.

## Usage

Inside of a git repository, run the following:

```bash
$> inertia init

$> inertia remote add glcoud 35.227.171.49 -u root -i /path/to/my/.ssh/id_rsa
Remote 'glcoud' added.

$> inertia gcloud init
Bootstrapping remote
Installing docker
Starting daemon
Building deploy key

Fetching daemon API token
Daemon running on instance
GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/keys/new):
ssh-rsa <...>

GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/hooks/new):
http://myhost.com:8081
Github WebHook Secret: inertia

Inertia daemon successfully deployed, add webhook url and deploy key to enable it.
Then run 'inertia gcloud up' to deploy your application.

$> inertia remote status gcloud
Remote instance 'gcloud' accepting requests at http://myhost.com:8081
```

A daemon is now running on your remote instance - but your application is not yet
continuously deployed.

The output of `inertia [REMOTE] init` has given you two important pieces of information:

1. A deploy key. The Inertia daemon requires readonly access to your GitHub repository.
   Add it to your GitHub repository settings at the URL provided in the output.
2. A GitHub webhook URL. The daemon will accept POST requests from GitHub at the URL
   provided. Again, add this webhook URL in your GitHub settings area (at the URL
   provided).

After adding these pieces of information to your GitHub settings,

```bash
$> inertia gcloud up
(Status code 201) Project up

$> inertia gcloud status
(Status code 200) 7b7be0b7097a26169e17037f4220fd0ce039bde1 refs/heads/master
Active containers:
project_frontend (/project_frontend_1)
project_web (/project_web_1)
project_solr (/project_solr_1)
postgres (/project_db_1)

$> inertia gcloud down
(Status code 200) Project down
```

## Development

### Dependencies

We use [dep](https://github.com/golang/dep) for managing dependencies. Install
that first if you haven't already.

```
brew install dep
```

Install project dependencies.

```bash
dep ensure
```

### Bootstrapping

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

### Testing

```bash
go test ./cmd -cover
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

## Setup

A primary design goal of Inertia is to minimize setup time for new projects. The
current setup flow is:

* Install and run Inertia on a new server
* Inertia will generate:
  * A SSH public key
  * A webhook URL and secret
* Add the SSH key to your project's Deploy Keys on GitHub
* Create a webhook with the URL and secret on your project repository
