---
title: API Reference

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
of project types on any virtual private server. The project is used, built, and maintained with ❤️ by [UBC Launch Pad](https://www.ubclaunchpad.com), UBC's
student-run software engineering club.

## Why use Inertia?

TODO

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

All you need to get started is a 
[compatible project](https://github.com/ubclaunchpad/inertia/wiki/Project-Compatibility), the Inertia CLI, and access to a virtual private server. 

For other platforms, you can [download the appropriate binary from the Releases page](https://github.com/ubclaunchpad/inertia/releases).

## Setup

```shell
inertia init
```

> This will generate a configuration file inside your repository.

<aside class="warning">
Do not commit the generated configuration file - add it to your 
<code>.gitignore</code>.
</aside>

## Configuring Inertia

```toml
version = "test"
project-name = "inertia"
build-type = "dockerfile"
build-file-path = "Dockerfile"

[remotes]
  [remotes.local]
    name = "local"
    IP = "127.0.0.1"
    user = "root"
    pemfile = "/Users/robertlin/.ssh/id_rsa"
    branch = "cmd/#504-cli-structure"
    ssh-port = "69"
    [remotes.local.daemon]
      port = "4303"
      token = "12345"
      webhook-secret = "abcde"
```

Your Inertia configuration contains a few key pieces of information.

# Deploying Your Project

## Using an Existing Remote

## Provisioning a Remote

```shell
inertia provision ec2 [remote_name]
```

## Configuring Your Repository

Inertia outputs a 
