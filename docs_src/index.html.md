---
title: Inertia Usage Guide

language_tabs: # must be one of https://git.io/vQNgJ
  # - shell

toc_footers:
  - <a href='https://github.com/ubclaunchpad/inertia'>GitHub Repository</a>
  - <a href='https://github.com/ubclaunchpad/inertia/issues/new/choose'>Report a Problem</a>
  - <a href='https://www.ubclaunchpad.com'><br/>UBC Launch Pad</a> 

includes:
  # - errors

search: false
---

# Inertia

✈️ Effortless, self-hosted continuous deployment for small teams and projects

[![](https://img.shields.io/github/release/ubclaunchpad/inertia.svg)](https://github.com/ubclaunchpad/inertia/releases/latest)
[![](https://img.shields.io/docker/pulls/ubclaunchpad/inertia.svg?colorB=0db7ed)](https://cloud.docker.com/u/ubclaunchpad/repository/docker/ubclaunchpad/inertia/general)
[![](https://img.shields.io/github/stars/ubclaunchpad/inertia.svg?style=social)](https://github.com/ubclaunchpad/inertia)

> **Main Features**
> 
> * 🚀 **Simple to use** - set up a deployment from your computer without ever having to manually SSH into your remote
> * 🍰 **Cloud-agnostic** - use any Linux-based remote virtual private server provider you want
> * ⚒  **Versatile project support** - deploy any Dockerfile or docker-compose project
> * 🚄 **Continuous deployment** - Webhook integrations for GitHub, GitLab, and Bitbucket means your project can be automatically updated, rebuilt, and deployed as soon as you `git push`
> * 🛂 **In-depth controls** - start up, shut down, and monitor your deployment with ease from the command line or using Inertia's REST API
> * 🏷 **Flexible configuration** - branch deployment, environment variables, easy file transfer for configuration files, build settings, and more
> * 📦 **Built-in provisioning** - easily provision and set up VPS instances for your project with supported providers such as Amazon Web Services using a single command
> * 👥 **Built for teams** - provide shared access to an Inertia deployment by adding users
> * 🔑 **Secure** - secured with access tokens and HTTPS across the board, as well as features like 2FA for user logins

<br />

*Inertia* is a simple cross-platform command line application that enables quick
and easy setup and management of continuous, automated deployment of a variety
of project types on any virtual private server. The project is used, built, and
maintained with ❤️ by [UBC Launch Pad](https://www.ubclaunchpad.com), UBC's
student-run software engineering club.

[UBC Launch Pad](https://www.ubclaunchpad.com) is a student-run software
engineering club at the University of British Columbia that aims to provide
students with a community where they can work together to build a all sorts of
cool projects, ranging from mobile apps and web services to cryptocurrencies and
machine learning applications.

Many of our projects rely on hosting providers for deployment. Unfortunately we
frequently change hosting providers based on available funding and sponsorship,
meaning our projects often need to be redeployed. On top of that, deployment
itself can already be a frustrating task, especially for students with little to
no experience setting up applications on remote hosts. Inertia is a project we
started to address these problems.

<br />

This site primarily documents how to set up and use Inertia - to learn more
about the project, check out our [GitHub repository](https://github.com/ubclaunchpad/inertia)!

<br />

# Getting Started

## Installation

> MacOS users can install the CLI using [Homebrew](https://brew.sh):

```shell
brew install ubclaunchpad/tap/inertia
```

> Windows users can install the CLI using [Scoop](http://scoop.sh):

```shell
scoop bucket add ubclaunchpad https://github.com/ubclaunchpad/scoop-bucket
scoop install inertia
```

> To build and install the CLI from source:

```shell
go get -u github.com/ubclaunchpad/inertia
```

The Inertia command line interface (CLI) can be installed from a few package
managers such as Homebrew and Scoop. For other platforms, you can 
[download the appropriate binary from the Releases page](https://github.com/ubclaunchpad/inertia/releases).

You can also build Inertia from source, though this requires Golang to be
installed.

To verify your installation, try running `inertia --help` - this should output
some helpful text about Inertia.

Run `inertia --version` to verify that you have the
[latest release](https://github.com/ubclaunchpad/inertia/releases/latest).

## Setup

> This will generate a configuration file inside your repository:

```
cd /path/to/project
inertia init
```

To set up Inertia, you must first navigate to your project directory, which
must be a git repository. If Inertia cannot detect your project type, it will
prompt for more information.

<aside class="warning">
<b>Do not commit the generated configuration file</b> - add it to your
<code>.gitignore</code>!
</aside>

## Project Configuration

> An example `inertia.toml`:

```toml
version = "test"
project-name = "my_project"
build-type = "dockerfile"
build-file-path = "dockerfiles/Dockerfile.web"

# ... other stuff
```

> To change a setting, you can edit the configuration file directly, or run:

```shell
inertia config set ${parameter} ${new_value}
```

Your Inertia configuration is stored in `inertia.toml` by default. There are
a few project-wide settings stored here:

Parameter         | Description
----------------- | -----------
`version`         | This should match the version of your Inertia CLI, which you can see by running `inertia --version`. It is used to determine which version of the [Inertia daemon](https://cloud.docker.com/u/ubclaunchpad/repository/docker/ubclaunchpad/inertia/) to use.
`project-name`    | The name of the project you are deploying.
`build-type`      | This should be either `dockerfile` or `docker-compose`, depending on which you are using.
`build-file-path` | Path to your build configuration file, such as `Dockerfile` or `docker-compose.yml`, relative to the root of your project.

# Deploying Your Project

When deploying a project, you typically deploy to a "remote".

A "remote" is a remote VPS, such as a [Google Cloud Compute](https://cloud.google.com/compute/)
or [AWS Elastic Cloud Compute (EC2)](https://aws.amazon.com/ec2/) instance. These
are computers in the cloud that will be used to deploy your project, so that
you don't have to use your own.

If this is your first time setting up a VPS, jump ahead to the
[Provisioning a Remote](#provisioning-a-remote) section, which will help you set
up a VPS for Inertia.

## Using an Existing Remote

> This command will prompt you for the path to your PEM file, your username,
> and the IP address of your remote. These parameters will be used to execute
> SSH commands that set up Inertia on your VPS.

```shell
inertia remote add ${remote_name}
```

To use an existing remote, you'll need its address and a PEM key that can be
used to access it. Inertia will also need a few ports exposed, namely one for
the Inertia daemon (port `4303` by default) and whatever ports you need for your
deployed project.

<aside class="notice">
If you use a non-standard SSH port (i.e. not port <code>22</code>) or want to
use a different port for the Inertia daemon, use the <code>--ssh.port ${port}</code>
and <code>--port ${port}</code> flags respectively when adding your remote.
</aside>

## Provisioning a Remote

```shell
inertia provision ${cloud_provider} ${remote_name}
```

Inertia has integrations with some cloud providers to allow you to easily
provision a new VPS instance and set it up for Inertia. You can run `inertia
provision --help` to see what options are available.

### Example: Provisioning an EC2 Instance

```shell
inertia provision ec2 my_remote \
  --from-profile \
  --ports 8080
```

> This command says: "provision an ec2 instance called 'my_remote' using
> credentials from my AWS profile and expose port 8080 for my project".

TODO: IAM setup, saving profile in aws config, choosing an image etc.

## Deployment Configuration

> An example `inertia.toml`:

```toml
# ... other stuff

[remotes]
  [remotes.my_remote]
    IP = "ec2-203-0-113-25.compute-1.amazonaws.com"
    user = "root"
    pemfile = "/Users/robertlin/.ssh/id_rsa"
    branch = "master"
    ssh-port = "22"
    [remotes.my_remote.daemon]
      port = "4303"
      token = ""
      webhook-secret = "abcdefg"
```

> To change a setting, you can edit the configuration file directly, or run:

```shell
inertia remote set ${remote_name} ${parameter} ${new_value}
```

> For example, the following will change the `branch` deployed on `my_remote`
> to `dev` and print out the new settings:

```shell
inertia remote set my_remote branch dev
inertia remote show my_remote
```

Once you've added a remote, remote-specific settings are available under the
`[remote]` section of your Inertia configuration. 

<aside class="notice">
For the most part, unless you filled in something incorrectly while adding a
remote or provisioning an instance, you won't need to change any of these
settings.
</aside>

Parameter  | Description
---------- | -----------
`IP`       | This is the address of your remote instance. It's how other users will access your deployed project as well!
`user`     | The user to use to execute commands as on your remote instance.
`pemfile`  | The key to use when executing SSH commands on your remote instance.
`branch`   | The git branch of your project that you want to deploy.
`ssh-port` | The SSH port on your remote instance - you usually don't need to change this.

Under `remotes.${remote_name}.daemon` there are some additional settings for the
Inertia daemon:

Parameter        | Description
---------------- | -----------
`port`           | The port that the Inertia daemon is using - you can usually leave this as is.
`token`          | This is the token used to authenticate against your remote, and will be populated when you initialize the Inertia daemon later.
`webhook-secret` | This is used to verify that incoming webhooks are authenticate - you'll need this later!

## Initializing the Inertia Daemon

<aside class="notice">
If you used <code>inertia provision</code> to set up your remote, you can skip
this step, as Inertia will have already done all this for you!
</aside>

```shell
inertia ${remote_name} init
# ... lots of output
```

Initializing the Inertia daemon means installing [Docker](https://www.docker.com/),
setting up some prerequisites, downloading the Inertia daemon image, and getting
it up and running. Luckily, this is all done by a single, handy command so you
don't have to worry about it!

The Inertia daemon is a small agent that runs on your VPS instance and handles
all your deployment-related needs, such as responding to commands and listening
for updates to your repository. You can read more about this in the
[Deployment Management](#deployment-management) section.

<aside class="success">
Once this step is done, it means that Inertia and all its prerequisites are now
set up on your VPS, and the Inertia daemon should be up and running! Try
running <code>inertia ${remote_name} status</code> to connect to the daemon.
</aside>

<aside class="notice">
This does <b>not</b> completely set up Inertia for project deployment - check
the next step to see what else needs to be done!
</aside>


## Configuring Your Repository

> The `inertia ${remote_name} init` or `provision` command's output should
> include something like the following:

```shell
GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/keys/new):
ssh-rsa <...> # this is important!
GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/hooks/new):
http://myhost.com:4303/webhook # this is important!
Github WebHook Secret: inertia # this is important!
```

The `inertia ${remote_name} init` or `provision` command outputs several key
pieces of information to get your repository set up for continuous deployment:

* A deploy key: you need to allow this key read access to your repository. On
  GitHub, this is under your project's "Settings -> Deploy Keys" tab. This will
  allow the Inertia daemon to clone your project.
* A webhook URL and webhook secret: you'll need to register the Inertia daemon
  for webhook updates to let it automatically deploy your latest changes. On
  GitHub, this is under your project's "Settings -> Webhooks" tab.

<aside class="warning">
Unless you've set up a custom SSL certificate for your remote, you'll have to
<b>disable SSL verification</b> when setting up your webhook registration.
</aside>

```shell
inertia ${remote_name} up
```

<aside class="success">
With your repository now configured correctly, you can now start get your project
running using the <code>up</code> command!
</aside>

# Deployment Management

> To bring your project online and check its status:

```shell
inertia ${remote_name} up
inertia ${remote_name} status
```

> To shut down your project:

```shell
inertia ${remote_name} down
```

The main commands used to control your deployed project are `up`, `status`, and
`down`. These commands are associated with the state of your project, not with
that of the Inertia daemon - `inertia ${remote_name} down`, for example, will
*not* shut down the Inertia daemon.

Commands like `up` will provide live output to your terminal. It's pretty normal
for `up` to take a while, depending on the performance of your VPS, as it needs
some time to build your project.

## Monitoring

```shell
inertia ${remote_name} logs
```

> To view the logs of a specific process, get the list of active containers, then
> query for the logs of the container you are interested in:

```shell
inertia ${remote_name} status
inertia ${remote_name} logs ${container_name}
```

TODO: details

## Secrets Management

> Environment variables are a good way to store secrets:

```shell
inertia ${remote_name} env set ${key} ${value}
```

> If you use configuration files such as a `.env` file, you can "send" it to your
> remote - this file will then become accessible by your project:

```shell
inertia ${remote_name} send ${file_name}
```

TODO: details

## TODO

Lorem ipsum

# Teams

## Configuring Users

> The following command will prompt for a password, and add the given user as
> an administrator:

```shell
inertia ${remote_name} user add ${username} --admin
```

> To list existing users:

```shell
inertia ${remote_name} user ls
```

> Access can be revoked for a user by removing them:

```shell
inertia ${remote_name} user rm ${username}
```

TODO

## Logging In

> If you want to log in to a remote you have already configured as a specific
> user, you can run:

```shell
inertia ${remote_name} login ${username}
```

TODO

<aside class="notice">
User tokens expire periodically for security, so you may have to log in again
from time to time.
</aside>

# Upgrading

> Install the latest release - for example, on MacOS:

```shell
brew upgrade ubclaunchpad/tap/inertia
inertia --version # verify installation
```

> To update configuration and daemon to match CLI version:

```shell
inertia config upgrade
inertia ${remote_name} upgrade
```

TODO

# Advanced Usage

TODO

## Troubleshooting

> To view more of the Inertia daemon's logs (the number of entries retrieved
> is capped by default), run:

```shell
inertia ${remote_name} logs --entries 100000
```

> To start an SSH session with your remote, you can use the shortcut:

```shell
inertia ${remote_name} ssh
```

TODO

## 2-Factor Authentication

```shell
inertia ${remote_name} user totp enable ${username}
inertia ${remote_name} user totp disable ${username}
```

2-factor authentication configuration in Inertia is available under the
`inertia ${remote_name} user totp` commands.

TOTP stands for "time-based one-time password". Enabling TOTP on an account
means that whenever a user logs in using that account, they must also verify
their identity using a separate authenticator app, which provides a secondary
factor of authentication - hence why it is called "2-factor authentication".

There are a wide range of authenticator apps out there, for example:

* [Authy for iOS](https://itunes.apple.com/ca/app/authy/id494168017?mt=8)
* [Authy for Android](https://play.google.com/store/apps/details?id=com.authy.authy&hl=en)

> To log in as a totp-enabled user, you'll have to provide your TOTP or backup
> code as a flag:

```shell
inertia ${remote_name} user login ${username} --totp ${code}
```

When you enable TOTP on an account, Inertia will output a QR code that you can
scan using your authenticator app, as well as a list of backup codes you should
keep somewhere safe. When you log in using a TOTP-enabled account, you'll need
to provide the TOTP generated by your authenticator app to log in, or one of the
backup codes.

## Resource Management

> To clear out unused Docker images and containers:

```shell
inertia ${remote_name} prune
```

> You can also interact with Docker over SSH:

```shell
inertia ${remote_name} ssh
root@remote:~$ docker container ls -a
```

Inertia does its best to manage resources on its own, but sometimes spare
containers and images start gathering and you might find your remote starting
to run out of storage (especially if it is a free/budget instance with low
storage).

Inertia offers a few ways of managing resources, either through commands like
`prune` or directly over SSH.

## Generating API Keys

```shell
inertia ${remote_name} token
```

If you want to develop integrations with Inertia, you'll probably want a
non-expiring API key, which you can generate using Inertia if you have SSH
access to the remote. Be careful not to lose these keys.

# Miscellaneous

## Learn More

> Some quick links to help you get started:

> * [Motivation and Design](https://github.com/ubclaunchpad/inertia#bulb-motivation-and-design)
> * [Continuous Deployment?](https://github.com/ubclaunchpad/inertia/wiki/Continuous-Deployment%3F)
> * [Architecture](https://github.com/ubclaunchpad/inertia/wiki/Architecture)
> * [Daemon, API, and Builds](https://github.com/ubclaunchpad/inertia/wiki/Daemon,-API,-and-Builds)
> * [Web App](https://github.com/ubclaunchpad/inertia/wiki/Web-App)
> * [CI Pipeline](https://github.com/ubclaunchpad/inertia/wiki/CI-Pipeline)
> * [Documentation](https://godoc.org/github.com/ubclaunchpad/inertia)

Check out our [GitHub repository](https://github.com/ubclaunchpad/inertia) and
[Wiki](https://github.com/ubclaunchpad/inertia/wiki)!

![](https://bobheadxi.github.io/assets/images/posts/inertia-diagram.png)

## Swag

```markdown
[![Deployed with Inertia](https://img.shields.io/badge/deploying%20with-inertia-blue.svg)](https://github.com/ubclaunchpad/inertia)
```

Add a cool Inertia badge to your README if you use Inertia!

[![Deployed with Inertia](https://img.shields.io/badge/deploying%20with-inertia-blue.svg)](https://github.com/ubclaunchpad/inertia)

# Contributing

Any contribution (pull requests, feedback, bug reports, ideas, etc.) is welcome!

You can report issues with documentation, bugs, or anything via our
[issue tracker](https://github.com/ubclaunchpad/inertia/issues).

For development, please see our
[contribution guide](https://github.com/ubclaunchpad/inertia/blob/master/CONTRIBUTING.md)
for contribution guidelines as well as a detailed guide to help you get started
with Inertia's codebase.

[![](https://golang.org/doc/gopher/pencil/gophermega.jpg)](https://golang.org/doc/gopher/pencil/)

<br />
<br />
<br />
<br />
<br />
