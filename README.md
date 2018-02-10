<p>
  <h1 align="center"> 👩‍🚀 Inertia </h1>
</p>

<p align="center">
  Simple, self-hosted continuous deployment.
</p>

<p align="center">
  <a href="https://travis-ci.org/ubclaunchpad/inertia">
    <img src="https://travis-ci.org/ubclaunchpad/inertia.svg?branch=master"
      alt="Built Status" />
  </a>

  <a href="https://coveralls.io/github/ubclaunchpad/inertia?branch=master">
    <img src="https://coveralls.io/repos/github/ubclaunchpad/inertia/badge.svg?branch=master"
      alt="Coverage Status" />
  </a>

  <a href="https://goreportcard.com/report/github.com/ubclaunchpad/inertia">
    <img src="https://goreportcard.com/badge/github.com/ubclaunchpad/inertia" alt="Clean code" />
  </a>

  <a href="">
    <img src="https://img.shields.io/badge/Shipping_faster_with-ZenHub-5e60ba.svg?style=flat" alt="We use Zenhub!" />
  </a>
</p>

<p align="center"> 
  <img src="https://img.shields.io/badge/Supported%20VPS%20platforms-Ubuntu%2014.04%2F16.04%20%7C%20CentOS%207-blue.svg" />
</p>

----------------

Inertia is a cross-platform command line tool that aims to simplify setup and management of automated deployment for docker-compose projects on any virtual private server.

## Installation

```bash
go get -u github.com/ubclaunchpad/inertia
```

Alternatively, you can download Inertia executables from the [Releases](https://github.com/ubclaunchpad/inertia/releases) page.

## Usage

Inside of a git repository, simply running the following commands to initialize Inertia and add a remote VPS:

```bash
$> inertia init

$> inertia remote add gcloud
Enter location of PEM file (leave blank to use '/Users/yourspecialname/.ssh/id_rsa'):
/path/to/my/.ssh/id_rsa
Enter user:
root
Enter IP address of remote:
35.227.171.49
Port 8081 will be used as the daemon port.
Run this 'inertia remote add' with the -p flag to set a custom port.

Remote 'gcloud' has been added!
You can now run 'inertia gcloud init' to set this remote up
for continuous deployment.
```

After adding a remote, you can now bring the Inertia daemon online:

```bash
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

A daemon is now running on your remote instance - but your application is not yet continuously deployed.

The output of `inertia [REMOTE] init` has given you two important pieces of information:

1. A deploy key. The Inertia daemon requires readonly access to your GitHub repository. Add it to your GitHub repository settings at the URL provided in the output.
2. A GitHub webhook URL. The daemon will accept POST requests from GitHub at the URL provided. Again, add this webhook URL in your GitHub settings area (at the URL provided).

After adding these pieces of information to your GitHub settings, the Inertia daemon will automatically deploy any changes you make to your repository's default branch. You can also manually manage your project's deployment through the CLI:

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

We use [dep](https://github.com/golang/dep) for managing dependencies.

```bash
$> brew install dep
$> dep ensure
```

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

### Testing

```bash
$> make test                              # test against ubuntu:latest
$> make test VPS_OS=ubuntu VERSION=14:04  # test against ubuntu:14.04
```

You can also start a container that sets up a mock VPS for testing:

```bash
$> go install
$> make testenv-ubuntu
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

### Motivation and Design

At Launch Pad we are frequently changing hosting providers based on available funding and sponsorship. Inertia is a project to develop an in-house continuous deployment system to make deploying applications simple and painless, regardless of the hosting provider.

Inertia contains two major components:

* Deployment daemon
* Command line interface

The deployment daemon will run persistently in the background, receiving webhook events from GitHub whenever new commits are pushed. The CLI will provide an interface to adjust settings, add repositories, etc.

This design differs from other similar tools because Inertia runs on the same server as the project it is deploying.

Another primary design goal of Inertia is to minimize setup time for new projects and maximize compatibility across different client and VPS platforms.
