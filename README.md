<p align="center">
  <img src="/.static/inertia-with-name.png" width="30%"/>
</p>

<p align="center">
  Simple, self-hosted continuous deployment.
</p>

<p align="center">
  <a href="https://travis-ci.org/ubclaunchpad/inertia">
    <img src="https://travis-ci.org/ubclaunchpad/inertia.svg?branch=master"
      alt="Built Status" />
  </a>

  <a href="https://goreportcard.com/report/github.com/ubclaunchpad/inertia">
    <img src="https://goreportcard.com/badge/github.com/ubclaunchpad/inertia" alt="Clean code" />
  </a>

  <a href="https://www.zenhub.com">
    <img src="https://img.shields.io/badge/Shipping_faster_with-ZenHub-5e60ba.svg?style=flat" alt="We use Zenhub!" />
  </a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/VPS%20Platforms-Ubuntu%2016.04%2F14.04%20%7C%20Debian%209.3%2F8%20%7C%20CentOS%207-blue.svg" />
</p>

----------------

Inertia is a cross-platform command line tool that aims to simplify setup and management of automated deployment of docker-compose projects on any virtual private server. It aims to provide the ease and flexibility of services like Heroku without the complexity of Kubernetes while still giving users full control over their projects.

## Installation

Inertia executables can be downloaded from the [Releases](https://github.com/ubclaunchpad/inertia/releases) page. You can then add the binary to your PATH or run it directly.

Alternatively, you can [build Inertia from source](https://github.com/ubclaunchpad/inertia#building-from-source).

## Usage

Initializing a project for use with Inertia only takes a few simple steps:

```bash
$> inertia init
$> inertia remote add $VPS_NAME
```

After adding a remote, you can bring the Inertia daemon online on your VPS:

```bash
$> inertia $VPS_NAME init
$> inertia remote status $VPS_NAME
# Confirms that the daemon is online and accepting requests
```

A daemon is now running on your remote instance - but your application is not yet continuously deployed. The output of `inertia [REMOTE] init` has given you two important pieces of information:

1. A deploy key:

```bash
GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/keys/new):
ssh-rsa <...>
```

The Inertia daemon requires readonly access to your GitHub repository. Add the deploy key to your GitHub repository settings at the URL provided in the output - this will grant the daemon access to clone your repository.

2. A GitHub webhook URL:

```bash
GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/hooks/new):
http://myhost.com:8081
Github WebHook Secret: inertia
``` 

The daemon will accept POST requests from GitHub at the URL provided. Add this webhook URL in your GitHub settings area (at the URL provided) so that the daemon will receive updates from GitHub when your repository is updated.

After adding these pieces of information to your GitHub settings, the Inertia daemon will automatically deploy any changes you make to your repository's default branch!

You can also manually manage your deployment through a variety of commands:

```bash
$> inertia $VPS_NAME --help
```

😎 [![Deployed with Inertia](https://img.shields.io/badge/Deploying%20with-Inertia-blue.svg)](https://github.com/ubclaunchpad/inertia) 😎 


```
[![Deployed with Inertia](https://img.shields.io/badge/Deploying%20with-Inertia-blue.svg)](https://github.com/ubclaunchpad/inertia)
```

## Development

### Building from Source

```bash
$ go get -u github.com/ubclaunchpad/inertia
```

It is highly recommended that you use a [tagged build](https://github.com/ubclaunchpad/inertia/releases) to ensure compatibility between the CLI and your Inertia daemon.

```bash
$ git checkout v0.1.0
$ make inertia-tagged
$ inertia --version
```

Alternatively, you can manually edit `.inertia.toml` to use your desired version.

### Dependencies

We use [dep](https://github.com/golang/dep) for managing dependencies.

```bash
$> go get -u github.com/golang/dep/cmd/dep
$> dep ensure
```

### Testing

```bash
$> make test                              # test against ubuntu:latest
$> make test VPS_OS=ubuntu VERSION=14.04  # test against ubuntu:14.04
```

You can also start a container that sets up a mock VPS for testing:

```bash
$> go install
$> make testenv VPS_OS=ubuntu VERSION=16.04
# defaults to ubuntu:lastest without args
# note the location of the key that is printed
```

You can treat this container just as you would treat a real VPS:

```bash
$> cd /path/to/my/dockercompose/project
$> inertia init
$> inertia remote add local
# PEM file: /test_env/test_key, User: 'root', Address: 0.0.0.0
$> inertia local init
$> inertia remote status local
Remote instance 'local' accepting requests at http://0.0.0.0:8081
```

If you run into this error when deploying onto the `testvps`:

```bash
docker: Error response from daemon: error creating aufs mount to /var/lib/docker/aufs/mnt/fed036790dfcc73da5f7c74a7264e617a2889ccf06f61dc4d426cf606de2f374-init: invalid argument.
```

You probably need to go into your Docker settings and add this line to the Docker daemon configuration file:

```js
{
  ...
  "storage-driver" : "aufs"
}
```

This sneaky configuration file can be found under `Docker -> Preferences -> Daemon -> Advanced -> Edit File`.

### Compiling Bash Scripts

To bootstrap servers, some bash scripting is often involved, but we'd like to avoid shipping bash scripts with our go binary. So we use [go-bindata](https://github.com/jteeuwen/go-bindata) to compile shell scripts into our go executables.

```bash
$> go get -u github.com/jteeuwen/go-bindata/...
```

If you make changes to the bootstrapping shell scripts in
`client/bootstrap/`, convert them to `Assets` by running:

```bash
$> make bootstrap
```

Then use your asset!

```go
shellScriptData, err := Asset("cmd/bootstrap/myshellscript.sh")
if err != nil {
  log.Fatal("No asset with that name")
}

// Optionally run shell script over SSH.
result, _ := remote.RunSSHCommand(string(shellScriptData))
```

### Motivation and Design

At Launch Pad we are frequently changing hosting providers based on available funding and sponsorship. Inertia is a project to develop an in-house continuous deployment system to make deploying applications simple and painless, regardless of the hosting provider.

Inertia contains two major components:

* Deployment daemon
* Command line interface

The deployment daemon will run persistently in the background, receiving webhook events from GitHub whenever new commits are pushed. The CLI will provide an interface to adjust settings, add repositories, etc.

This design differs from other similar tools because Inertia runs on the same server as the project it is deploying.

Another primary design goal of Inertia is to minimize setup time for new projects and maximize compatibility across different client and VPS platforms.
