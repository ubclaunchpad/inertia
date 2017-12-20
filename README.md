# 👩‍🚀 Inertia

Inertia makes it easy to set up automated deployment for Dockerized
applications.

```bash
go get -u github.com/ubclauncpad/inertia
```

## Deploy an Application

Applications are deployed over SSH. You will need an SSH username and PEM file
to get started. Inside of a git repository, run the following:

```bash
$> inertia init

$> inertia remote add glcoud 35.227.171.49 -u root -i /path/to/my/.ssh/id_rsa
Remote 'glcoud' added.

$> inertia deploy
Deploying remote...
Daemon running on instance
GitHub Deploy Key Generation:
Generating public/private rsa key pair.
Your identification has been saved in /home/root/.ssh/id_rsa_inertia_deploy.
Your public key has been saved in /home/root/.ssh/id_rsa_inertia_deploy.pub.
The key fingerprint is:
SHA256:EO6Wp6QkeDPf67ODy5W329bJiEZcHKSVBRYZ0BKbFPU root@instance
The keys randomart image is:
+---[RSA 2048]----+
|      . =BOB.    |
|     . o.*=.     |
|      o +o .E    |
| .   . o  o      |
|. = . =.S.       |
| . * = +o        |
|    o.=... + .   |
|   ...ooooo +    |
|    oo+=oo.      |
+----[SHA256]-----+
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCftKIy4/GQah6H4EcxdO5Qmdin6Xu/9DoBE7Qh1L1P44B08szTJkzjhcMNexr0bzLstU+nks8qQT66zfkfih89gFb+7kF4KsZT5ITMAO/gZyqCoAMS/1FxQVkLvcMrAxTbXOcU3Uvq39RN2ELec5I6AaVZe328495fuB2RyLehYcS0oEWd8+WVA/0iS+qHx7yKacdOFkmX7LZOrdY1F4IMJpN+t1/oiSaBF77b1Fjhvlw9/iOMkj2P1tUudsh5QhXCWWBO0FmzyvIgSWx24PmU7cL131Ok6KhDukv62YAZj0Vmk73bvMrma5DWqK35+FNUi0IMMKlV3X5JyDY4pRt9 root@instance

GitHub WebHook URL: 35.227.171.49:8081
```

A daemon is now running on your remote instance, continuously deploying your
application for you!

The output of `inertia deploy` has given you two important pieces of information.

1. A deploy key. The Inertia daemon requires readonly access to your GitHub repository.
   Add it to your GitHub repository settings at the URL provided in the output.
2. A GitHub webhook URL. The daemon will accept POST requests from GitHub at the URL
   provided. Again, add this webhook URL in your GitHub settings area (at the URL
   provided).

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

