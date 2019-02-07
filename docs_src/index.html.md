---
title: Inertia Usage Guide

language_tabs: # must be one of https://git.io/vQNgJ
  # - shell

toc_footers:
  - <a href='https://github.com/ubclaunchpad/inertia'>GitHub Repository</a>
  - <a href='https://github.com/ubclaunchpad/inertia/issues/new/choose'>Report a Problem</a>

includes:
  # - errors

search: true
---

# Inertia

✈️ Effortless, self-hosted continuous deployment for small teams and projects

[![](https://img.shields.io/github/release/ubclaunchpad/inertia.svg)](https://github.com/ubclaunchpad/inertia/releases/latest)
[![](https://img.shields.io/docker/pulls/ubclaunchpad/inertia.svg?colorB=0db7ed)](https://cloud.docker.com/u/ubclaunchpad/repository/docker/ubclaunchpad/inertia/general)
[![](https://img.shields.io/github/stars/ubclaunchpad/inertia.svg?style=social)](https://github.com/ubclaunchpad/inertia)

## What is Inertia?

Inertia is a simple cross-platform command line application that enables quick
and easy setup and management of continuous, automated deployment of a variety
of project types on any virtual private server. The project is used, built, and
maintained with ❤️ by [UBC Launch Pad](https://www.ubclaunchpad.com), UBC's
student-run software engineering club.

Check out our [GitHub repository](https://github.com/ubclaunchpad/inertia) to
learn more!

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

> Verify that Inertia has been installed properly:

```shell
# should output a version number correctly
inertia --version

# display help text
inertia --help
```

Inertia can be installed from a few package managers, such as Homebrew and Scoop.

For other platforms, you can [download the appropriate binary from the Releases page](https://github.com/ubclaunchpad/inertia/releases).

You can also build Inertia from source.

## Setup

```
cd /path/to/project
inertia init
```

To set up Inertia, you must first navigate to your project directory, which
must be a git repository. If Inertia cannot detect your project type, it will
prompt for more information.

> This will generate a configuration file inside your repository.

<aside class="warning">
Do not commit the generated configuration file - add it to your 
<code>.gitignore</code>.
</aside>

## Project Configuration

> An example `inertia.toml`:

```toml
version = "test"
project-name = "inertia"
build-type = "dockerfile"
build-file-path = "Dockerfile"
```

Your Inertia configuration, stored in `inertia.toml` by default, contains a few
key pieces of information.

Parameter | Default | Description
--------- | ------- | -----------
TODO | false | TODO
TODO | true | TODO

# Deploying Your Project

## Using an Existing Remote

```shell
inertia remote add ${remote_name}
```

To use an existing remote, you'll need its address and a PEM key that can
be used to access it. Inertia will also need a few ports exposed, namely
one for its daemon (`4303` by default) and whatever ports you need for your
deployed project.

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

TODO: IAM setup, saving profile in aws config, choosing an image etc.

## Deployment Configuration

> In your `inertia.toml`:

```toml
[remotes]
  [remotes.my_remote]
    name = "my_remote"
    IP = "127.0.0.1"
    user = "root"
    pemfile = "/Users/robertlin/.ssh/id_rsa"
    branch = "cmd/#504-cli-structure"
    ssh-port = "69"
    [remotes.local.daemon]
      port = "4303"
      token = "12345678910"
      webhook-secret = "abcdefg"
```

Deployment-specific settings are available under the `[remote]` section of your
Inertia configuration.

Parameter | Default | Description
--------- | ------- | -----------
TODO | false | TODO
TODO | true | TODO

## Configuring Your Repository

> The `inertia ${remote_name} init` command should output something like the
> following:

```sh
GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/keys/new):
ssh-rsa <...>
GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/hooks/new):
http://myhost.com:4303/webhook
Github WebHook Secret: inertia
```

TODO: deploy key, webhooks, etc.

> Add some Inertia bling to your project repository!

```markdown
[![Deployed with Inertia](https://img.shields.io/badge/deploying%20with-inertia-blue.svg)](https://github.com/ubclaunchpad/inertia)
```

# Managing Your Deployment

## TODO

Lorem ipsum

## TODO

Lorem ipsum