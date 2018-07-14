# :books: Contributing

This document outlines key considerations and tips for anyone contributing to Inertia.

- [Opening an Issue](#opening-an-issue)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Development Tips](#development-tips)

------

# Opening an Issue

🎉 An issue, whether it be bugs, feature requests, or general feedback is welcome!

However, please do a quick search of past issues before opening a ticket. If you are working on a ticket, please assign it to yourself or leave a comment noting that you are working on it - similarly, if you decide to stop working on a ticket before it gets resolved, please un-assign yourself or leave a comment. This helps us keep track of which tickets are in progress.

# Submitting a Pull Request

👍 Contributions of any size and scope are very much appreciated!

All pull requests should be connected to one or more issues. Please try to fill out the pull request template to the best of your ability and use a clear, descriptive title.

At the very least, all pull requests need to pass our Travis builds and receive an approval from a reviewer. Please include tests whenever possible.

See the [Development Tips](#development-tips) section to get started with the codebase!

## Guidelines

### Commits

When writing a commit message, consider how useful it might be to someone reading it - can a reader tell *why* you made your changes based on your commit message? Try to avoid messages like `fix` or `wip` - for example, write `Fix x` or `Scaffold y`.

Formatting-wise, just follow basic Git commit message conventions - capitalize subject line, use the imperative mood, and so on. See [this guide](https://chris.beams.io/posts/git-commit/#seven-rules) for an introduction on writing good Git commit messages.

### Merging Pull Requests

Small, nuclear changes should be squashed - on our ZenHub board, this usually means tickets with 3 or fewer Epic Points. Pull requests with lots of "mistake" commits (`add back accidentally deleted file` or `wip wip`) should be squashed as well. 

Larger pull requests, given the commits are reasonable, should be merged with a standard `merge` to preserve history. On our ZenHub board, this usually means tickets with 8 or more Epic Points.

### Branch Naming

Branches should be named to refer to the component of Inertia the changes pertain to, as well as a related ticket. The component should correspond to [Labels](https://github.com/ubclaunchpad/inertia/labels) that begin with `area: ...`. The format ideally goes:

```
[area]/#[ticket]-[summary]
```

For example, [Issue #261](https://github.com/ubclaunchpad/inertia/issues/261) has the label `area: client` - in that case, the branch name should be `client/#261-ec2-provisioning`. If there are multiple `area` labels, just choose the most relevant one.

# Development Tips

👷 This section will walk you through Inertia's codebase, how to get a development environment set up, and outline the various tools available to help you out.

Please free free to open up a ticket if any of these instructions are unclear or straight up do not work on your platform!

- [Installation and Setup](#installation-and-setup)
- [Project Overview](#project-overview)
- [Testing Environment](#setting-up-a-testing-environment)

## Setup

First, [install Go](https://golang.org/doc/install#install) and grab Inertia's source code:

```bash
$> go get -u github.com/ubclaunchpad/inertia
```

If you are looking to contribute, you can then set your own fork as a remote:

```bash
$> git remote rename origin upstream   # Set the official repo as you
                                       # "upstream" so you can pull
                                       # updates
$> git remote add origin https://github.com/$AMAZING_YOU/inertia.git
```

You will also want to add `GOPATH` and `GOBIN` to your `PATH` to use any Inertia executables you install. Just add the following to your `.bashrc` or `.bash_profile`:

```bash
export PATH="$PATH:$HOME/go/bin"
export GOPATH=$HOME/go
export GOBIN=$HOME/go/bin
```

Inertia uses:
- [dep](https://github.com/golang/dep) for managing Golang dependencies
- [npm](https://www.npmjs.com) to manage dependencies for Inertia's React web app
- [Docker](https://www.docker.com/community-edition) for various application functionalities and integration testing

Make sure all of the above are installed (and that the Docker daemon is online) before running:

```bash
$> make deps          # installs dependencies
$> make cli           # installs Inertia build tagged as "test" to gopath
$> inertia --version  # check what version you have installed
```

A build tagged as `test` allows you to use `make testdaemon` for local development. See the next section for more details. Alternatively, you can manually edit `.inertia.toml` to use your desired daemon version - see the [Release Streams](#release-streams) documentation for more details.

Note that if you install Inertia using these commands or any variation of `go install`, you may have to remove the binary using `go clean -i github.com/ubclaunchpad/inertia` to use an Inertia CLI installed using Homebrew. To go back to a `go install`ed version of Inertia, you need to run `brew uninstall inertia`.

## Project Overview

[![GoDoc](https://godoc.org/github.com/ubclaunchpad/inertia?status.svg)](https://godoc.org/github.com/ubclaunchpad/inertia)

The Inertia codebase is split up into several components - this section gives a quick introduction on how to work with each.

### CLI

Inertia's command line application is initiated in the root directory, but the majority of the code is in the `cmd` package. It is built on top of [cobra](https://github.com/spf13/cobra), a library for building command line applications.

This code should only include the CLI user interface and code used to manage local assets, such as configuration files - core client logic, functionality, and daemon API interactions should go into the `client` package.

### Client

The Inertia client package manages all clientside functionality. The client codebase is in `./client/`.

To bootstrap servers, some bash scripting is often involved, but we'd like to avoid shipping bash scripts with our go binary - instead, we use [fileb0x](https://github.com/UnnoTed/fileb0x) to compile shell scripts into our Go executables. If you make changes to the bootstrapping shell scripts in `client/scripts/`, compile them by running:

```bash
$> make scripts
```

Then use your asset!

```go
shellScriptData, err := ReadFile("client/scripts/myshellscript.sh")
if err != nil {
  log.Fatal("No asset with that name")
}

// Optionally run shell script over SSH.
result, _ := remote.RunSSHCommand(string(shellScriptData))
```

### Daemon

The Inertia daemon package manages all serverside functionality and is the core of the Inertia platform. The daemon codebase is in `./daemon/inertiad/`.

To locally test a daemon compiled from source, set your Inertia version in `.inertia.toml` to `test` and run:

```bash
$> make testdaemon
# In your test project directory:
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

### Web

Inertia Web provides a web interface to manage an Inertia deployment. The web application codebase is in `./daemon/web/`.

To run a local instance of Inertia Web:

```bash
$> make web-deps   # install npm dependencies
$> make web-run    # run local instance of application                    
```

Make sure you have a local daemon set up for this web app to work - see the previous section for more details.

## Setting up a Testing Environment

You will need Docker installed and running to run whole the Inertia test suite, which includes a number of integration tests.

```bash
$> make dev-deps                              # install various development dependencies
$> make test-all                              # test against ubuntu:latest
$> make test-all VPS_OS=ubuntu VERSION=14.04  # test against ubuntu:14.04
```

Alternatively, `make test` will just run the unit tests.

Setting up a more comprehensive test environment, where you take a project from setup to deployment using Inertia, is a bit trickier - these are the recommended steps:

1. **Manually set up a mock VPS**

```bash
$> make testenv VPS_OS=ubuntu VERSION=16.04
# This defaults to ubuntu:lastest without args.
# Note the location of the key that is printed and use that when
# adding your local remote.
```

You can [SSH into this testvps container](https://bobheadxi.github.io/dockerception/#ssh-services-in-docker) and otherwise treat it just as you would treat a real VPS.

2. **Compile and install Inertia**

```bash
$> make
```

3. **Build and deliver Inertia daemon to the `testvps`**

```bash
$> make testdaemon
```

4. **Set up a test project**

You will need a GitHub repository you own, since you need permission to add deploy keys. The Inertia team typically uses the [inertia-deploy-test](https://github.com/ubclaunchpad/inertia-deploy-test) repository - you could just fork this repository.

```bash
$> git clone https://github.com/$AWESOME_YOU/inertia-deploy-test.git
$> cd inertia-deploy-test
$> inertia init
$> inertia remote add local
# - PEM file: $INERTIA_PATH/test/keys/id_rsa
# - Address:  127.0.0.1 
# - User:     root
$> inertia local init
$> inertia local status
```

The above steps will pull and use a daemon image from Docker Hub based on the version in your `.inertia.toml`.

Following these steps, you can run Inertia through deployment:

```bash
$> inertia local up --stream
$> inertia local status
$> inertia local logs
```

Please free free to open up an Issue if any of these steps are not clear or don't work!
