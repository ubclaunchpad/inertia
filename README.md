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

  <a href="https://github.com/ubclaunchpad/inertia/releases">
    <img src="https://img.shields.io/github/release/ubclaunchpad/inertia.svg" />
  </a>
</p>

----------------

Inertia is a cross-platform command line tool that simplifies setup and management of automated project deployment on any virtual private server. It aims to provide the ease and flexibility of services like Heroku without the complexity of Kubernetes while still giving users full control over their projects.

- [Usage](#package-usage)
  - [Setup](#setup)
  - [Continuous Deployment](#continuous-deployment)
  - [Deployment Management](#deployment-management)
  - [Release Streams](#release-streams)
- [Motivation and Design](#bulb-motivation-and-design)
- [Development](#construction-development)

# :package: Usage

All you need to get started is a docker-compose project, the Inertia CLI, and access to a virtual private server. To download the CLI:

```bash
$> brew install ubclaunchpad/tap/inertia
$> inertia
```

For other platforms, you can also [download the appropriate binary from the Releases page](https://github.com/ubclaunchpad/inertia/releases) or [install Inertia from source](#installing-from-source).

## Setup

Initializing a project for use with Inertia only takes a few simple steps:

```bash
$> inertia init
$> inertia remote add $VPS_NAME
```

After adding a remote, you can bring the Inertia daemon online on your VPS:

```bash
$> inertia $VPS_NAME init
$> inertia $VPS_NAME status
# Confirms that the daemon is online and accepting requests
```

This daemon will be used to manage your deployment.

See our [wiki](https://github.com/ubclaunchpad/inertia/wiki/VPS-Compatibility) for more details on VPS platform compatibility.

## Deployment Management

To manually deploy your project, you must first grant Inertia permission to clone your repository. This can be done by adding the GitHub Deploy Key that is displayed in the output of `inertia $VPS_NAME init` to your repository settings:

```bash
GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/keys/new):
ssh-rsa <...>
```

Once this is done, you can use Inertia to bring your project online on your remote VPS:

```bash
$> inertia $VPS_NAME up --stream
```

Run `inertia $VPS_NAME --help` to see the other commands Inertia offers for managing your deployment.

## Continuous Deployment

To enable continuous deployment, you need the webhook URL that is printed during `inertia $VPS_NAME init`:

```bash
GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/hooks/new):
http://myhost.com:8081
Github WebHook Secret: inertia
``` 

The daemon will accept POST requests from GitHub at the URL provided. Add this webhook URL in your GitHub settings area (at the URL provided) so that the daemon will receive updates from GitHub when your repository is updated. Once this is done, the daemon will automatically build and deploy any changes that are made to the deployed branch.

## Release Streams

The version of Inertia you are using can be seen in Inertia's `.inertia.toml` configuration file, or by running `inertia --version`. The version in `.inertia.toml` is used to determine what version of the Inertia daemon to use when you run `inertia $VPS_NAME init`.

You can manually change the daemon version used by editing the Inertia configuration file. If you are building from source, you can also check out the desired version and run `make inertia-tagged` or `make RELEASE=$STREAM`. Inertia daemon releases are tagged as follows:

- `v0.x.x` denotes [official, tagged releases](https://github.com/ubclaunchpad/inertia/releases) - these are recommended.
- `latest` denotes the newest builds on `master`.
- `canary` denotes experimental builds used for testing and development - do not use this.

The daemon component of an Inertia release can be patched separately from the CLI component - see our [wiki](https://github.com/ubclaunchpad/inertia/wiki/Daemon-Releases) for more details.

## Swag

Add some bling to your Inertia-deployed project :sunglasses:

```
[![Deployed with Inertia](https://img.shields.io/badge/Deploying%20with-Inertia-blue.svg)](https://github.com/ubclaunchpad/inertia)
```

[![Deployed with Inertia](https://img.shields.io/badge/Deploying%20with-Inertia-blue.svg)](https://github.com/ubclaunchpad/inertia)

# :bulb: Motivation and Design

At [UBC Launch Pad](http://www.ubclaunchpad.com), we are frequently changing hosting providers based on available funding and sponsorship. Inertia is a project to develop an in-house continuous deployment system to make deploying applications simple and painless, regardless of the hosting provider.

The primary design goals of Inertia are to:

* minimize setup time for new projects
* maximimise compatibility across different client and VPS platforms
* offer a convenient interface for managing the deployed application

## How It Works

Inertia consists of two major components: a deployment daemon and a command line interface.

The deployment daemon runs persistently in the background on the server, receiving webhook events from GitHub whenever new commits are pushed. The CLI provides an interface to adjust settings and manage the deployment - this is done through requests to the daemon, authenticated using JSON web tokens generated by the daemon. Remote configuration is stored locally in `.inertia.toml`.

<p align="center">
<img src="https://bobheadxi.github.io/assets/images/posts/inertia-diagram.png" width="70%" />
</p>

Inertia is set up serverside by executing a script over SSH that installs Docker and starts an Inertia daemon image with [access to the host Docker socket](https://bobheadxi.github.io/dockerception/#docker-in-docker). This Docker-in-Docker configuration gives the daemon the ability to start up other containers *alongside* it, rather than *within* it, as required. Once the daemon is set up, we avoid using further SSH commands and execute Docker commands through Docker's Golang API. Instead of installing the docker-compose toolset, we [use a docker-compose image](https://bobheadxi.github.io/dockerception/#docker-compose-in-docker) to build and deploy user projects.

# :construction: Development

This section outlines the various tools available to help you get started developing Inertia. Run `make commands` to list all the Makefile shortcuts available.

If you would like to contribute, feel free to comment on an issue or make one and open up a pull request!

## Installing from Source

First, [install Go](https://golang.org/doc/install#install) and grab Inertia's source code:

```bash
$> go get -u github.com/ubclaunchpad/inertia
```

We use [dep](https://github.com/golang/dep) for managing Golang dependencies, and [npm](https://www.npmjs.com) to manage dependencies for Inertia's React web app. Make sure both are installed before running the following commands.

```bash
$> dep ensure     # Inertia CLI and daemon dependencies
$> make web-deps  # Inertia Web dependencies
```

For usage, it is highly recommended that you use a [tagged build](https://github.com/ubclaunchpad/inertia/releases) to ensure compatibility between the CLI and your Inertia daemon.

```bash
$> git checkout $VERSION
$> make inertia-tagged    # installs a tagged version of Inertia
$> inertia --version      # check what version you have installed
```

Alternatively, you can manually edit `.inertia.toml` to use your desired daemon version - see the [Release Streams](#release-streams) documentation for more details.

Note that if you install Inertia using these commands or any variation of `go install`, you may have to remove the binary using `go clean -i github.com/ubclaunchpad/inertia` to go back to using an Inertia CLI installed using Homebrew.

For development, you should install a build tagged as `test` so that you can make use `make testdaemon` for local development. See the next section for more details.

```bash
$> make RELEASE=test    # installs current Inertia build and mark as "test"
```

## Testing and Locally Deploying

You will need Docker installed and running to run the Inertia test suite, which includes a number of integration tests.

```bash
$> make test                              # test against ubuntu:latest
$> make test VPS_OS=ubuntu VERSION=14.04  # test against ubuntu:14.04
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
# PEM file: inertia/test_env/test_key, User: 'root', Address: 0.0.0.0
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
