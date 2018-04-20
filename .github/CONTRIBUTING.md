# Contributing

This document outlines key considerations and tips for anyone contributing to Inertia.

- [Opening an Issue](#opening-an-issue)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Development Tips](#development-tips)

------

# Opening an Issue

Please do a quick search of past issues before opening a ticket. If working on a ticket, please assign it to yourself or leave a comment saying you are working on it. If you have decide to stop working on a ticket before it gets resolved, please un-assign yourself.

# Submitting a Pull Request

All pull requests should be connected to one or more issues. Please try to fill out the pull request template to the best of your ability and use a clear, descriptive title.

At the very least, all pull requests need to pass our Travis builds and receive an approval from a reviewer. Please include tests whenever possible.

# Development Tips

This section outlines the various tools available to help you get started developing Inertia. Run `make ls` to list all the Makefile shortcuts available.

If you would like to contribute, feel free to comment on an issue or make one and open up a pull request!

## Setup

First, [install Go](https://golang.org/doc/install#install) and grab Inertia's source code:

```bash
$> go get -u github.com/ubclaunchpad/inertia
```

We use [dep](https://github.com/golang/dep) for managing Golang dependencies, and [npm](https://www.npmjs.com) to manage dependencies for Inertia's React web app. Make sure both are installed before running the following commands.

```bash
$> dep deps           # Install all dependencies
$> make RELEASE=test  # installs Inertia build tagged as "test"
$> inertia --version  # check what version you have installed
```

A build tagged as `test` allows you to use `make testdaemon` for local development. See the next section for more details.

Alternatively, you can manually edit `.inertia.toml` to use your desired daemon version - see the [Release Streams](#release-streams) documentation for more details.

Note that if you install Inertia using these commands or any variation of `go install`, you may have to remove the binary using `go clean -i github.com/ubclaunchpad/inertia` to go back to using an Inertia CLI installed using Homebrew. To go back to a `go install`ed version of Inertia, you need to run `brew uninstall inertia`.

## Repository Structure

The codebase for the CLI is in the root directory. This code should only include the user interface - all client-based logic and functionality should go into the client.

### Client

The Inertia client manages all clientside functionality. The client codebase is in `client/`.

### Daemon

The Inertia daemon manages all serverside functionality. The daemon codebase is in `daemon/inertia`.

### Inertia Web

The Inertia Web application provides a web interface to manage an Inertia deployment. The web application codebase is in `daemon/web`.

## Testing and Locally Deploying

You will need Docker installed and running to run the Inertia test suite, which includes a number of integration tests.

```bash
$> make test-all                              # test against ubuntu:latest
$> make test-all VPS_OS=ubuntu VERSION=14.04  # test against ubuntu:14.04
```

You can also manually start a container that sets up a mock VPS for testing:

```bash
$> make testenv VPS_OS=ubuntu VERSION=16.04
# This defaults to ubuntu:lastest without args.
# Note the location of the key that is printed and use that when
# adding your local remote.
```

You can [SSH into this testvps container](https://bobheadxi.github.io/dockerception/#ssh-services-in-docker) and otherwise treat it just as you would treat a real VPS:

```bash
$> cd /path/to/my/dockercompose/project
$> inertia init
$> inertia remote add local
# PEM file: inertia/test/keys/id_rsa, User: 'root', Address: 0.0.0.0
$> inertia local init
$> inertia local status
```

The above steps will pull and use a daemon image from Docker Hub based on the version in your `.inertia.toml`.

### Daemon

To use a daemon compiled from source, set your Inertia version in `.inertia.toml` to `test` and run:

```bash
$> make testdaemon
$> inertia local init
```

This will build a daemon image and `scp` it over to the test VPS, and use that image for the daemon when setting up `testvps` using `inertia local init`

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

### Web App

Inertia Web is a React application. To run a local instance of Inertia Web:

```bash
$> make web-run
```

Make sure you have a local daemon set up for this web app to work - see the previous section for more details.

## Compiling Bash Scripts

To bootstrap servers, some bash scripting is often involved, but we'd like to avoid shipping bash scripts with our go binary. So we use [go-bindata](https://github.com/jteeuwen/go-bindata) to compile shell scripts into our go executables.

```bash
$> go get -u github.com/jteeuwen/go-bindata/...
```

If you make changes to the bootstrapping shell scripts in `client/bootstrap/`, convert them to `Assets` by running:

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
