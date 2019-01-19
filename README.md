<p align="center">
  <img src="/.static/inertia-with-name.png" width="25%" alt="Inertia"/>
</p>

<p align="center">
  An effortless, self-hosted continuous deployment platform.
</p>

<p align="center">
  <a href="https://travis-ci.org/ubclaunchpad/inertia">
    <img src="https://travis-ci.org/ubclaunchpad/inertia.svg?branch=master"
      alt="Travis Build Status" />
  </a>

  <a href="https://ci.appveyor.com/project/ubclaunchpad/inertia">
    <img src="https://ci.appveyor.com/api/projects/status/2fll6p9677bujb7q/branch/master?svg=true"
      alt="Appveyor Build Status" />
  </a>

  <a href="https://goreportcard.com/report/github.com/ubclaunchpad/inertia">
    <img src="https://goreportcard.com/badge/github.com/ubclaunchpad/inertia"
      alt="Clean code" />
  </a>

  <a href="https://godoc.org/github.com/ubclaunchpad/inertia">
    <img src="https://godoc.org/github.com/ubclaunchpad/inertia?status.svg"
      alt="GoDocs available" />
  </a>

  <a href="https://cloud.docker.com/u/ubclaunchpad/repository/docker/ubclaunchpad/inertia/general">
    <img src="https://img.shields.io/docker/pulls/ubclaunchpad/inertia.svg?colorB=0db7ed"
      alt="inertiad image">
  </a>

  <a href="https://github.com/ubclaunchpad/inertia/releases/latest">
    <img src="https://img.shields.io/github/release/ubclaunchpad/inertia.svg"
      alt="Latest release" />
  </a>
</p>

<br>

<p align="center">
  <a href="#package-getting-started"><strong>Getting Started</strong></a> · 
  <a href="#bulb-motivation-and-design"><strong>Motivation & Design</strong></a> · 
  <a href="#books-contributing"><strong>Contributing</strong></a> · 
  <a href="https://github.com/ubclaunchpad/inertia/wiki"><strong>Wiki</strong></a>
</p>

<br>

<p align="center">
    <img src="/.static/inertia-init.png" width="35%" />
</p>

<br>

Inertia is a simple cross-platform command line application that enables quick and easy setup and management of continuous, automated deployment of a variety of project types on any virtual private server. The project is used, built, and maintained with :heart: by [UBC Launch Pad](https://www.ubclaunchpad.com/), UBC's student-run software engineering club.

|   | Main Features  |
----|-----------------
🚀  | **Simple to use** - set up a deployment from your computer without ever having to manually SSH into your remote
🍰  | **Cloud-agnostic** - use any Linux-based remote virtual private server provider you want
⚒  | **Versatile project support** - deploy any Dockerfile or docker-compose project
🚄  | **Continuous deployment** - Webhook integrations for GitHub, GitLab, and Bitbucket means your project can be automatically updated, rebuilt, and deployed as soon as you `git push`
🛂  | **In-depth controls** - start up, shut down, and monitor your deployment with ease from the command line or using Inertia's REST API
🏷  | **Flexible configuration** - branch deployment, environment variables, easy file transfer for configuration files, build settings, and more
📦  | **Built-in provisioning** - easily provision and set up VPS instances for your project with supported providers such as Amazon Web Services using a single command
👥  | **Built for teams** - provide shared access to an Inertia deployment by adding users
🔑  | **Secure** - secured with access tokens and HTTPS across the board, as well as features like 2FA for user logins

<br>

# :package: Getting Started

All you need to get started is a [compatible project](https://github.com/ubclaunchpad/inertia/wiki/Project-Compatibility), the Inertia CLI, and access to a virtual private server. 

MacOS users can install the CLI using [Homebrew](https://brew.sh):

```bash
$> brew install ubclaunchpad/tap/inertia
```

Windows users can install the CLI using [Scoop](http://scoop.sh):

```bash
$> scoop bucket add ubclaunchpad https://github.com/ubclaunchpad/scoop-bucket
$> scoop install inertia
```

For other platforms, you can [download the appropriate binary from the Releases page](https://github.com/ubclaunchpad/inertia/releases).

### Setup

Initializing a project for use with Inertia only takes a few simple steps:

```bash
$> inertia init
```

#### Using an Existing Remote

To use an existing host, you must first add it to your Inertia configuration and initialize it - this will install Inertia on your remote.

```bash
$> inertia remote add $VPS_NAME
$> inertia $VPS_NAME init
$> inertia $VPS_NAME status
# Confirms that the daemon is online and accepting requests
```

See our [wiki](https://github.com/ubclaunchpad/inertia/wiki/VPS-Compatibility) for more details on VPS platform compatibility.

#### Provisioning a New Remote

Inertia offers some tools to easily provision a new VPS instance and set it up for Inertia. For example, to create an EC2 instance and initialize it, just run:

```bash
$> inertia provision ec2 $VPS_NAME
$> inertia $VPS_NAME status
```

### Deployment Management

To manually deploy your project, you must first grant Inertia permission to clone your repository. This can be done by adding the GitHub Deploy Key that is displayed in the output of `inertia $VPS_NAME init` to your repository settings:

```bash
GitHub Deploy Key (add here https://www.github.com/<your_repo>/settings/keys/new):
ssh-rsa <...>
```

Once this is done, you can use Inertia to bring your project online on your remote VPS:

```bash
$> inertia $VPS_NAME up
```

Run `inertia $VPS_NAME --help` to see the other commands Inertia offers for managing your deployment.

Inertia also offers a web application - this can be accessed at `https://$ADDRESS:4303/web` once users have been added through the `inertia $VPS_NAME user` commands.

### Continuous Deployment

To enable continuous deployment, you need the webhook URL that is printed during `inertia $VPS_NAME init`:

```bash
GitHub WebHook URL (add here https://www.github.com/<your_repo>/settings/hooks/new):
http://myhost.com:4303/webhook
Github WebHook Secret: inertia
``` 

The daemon will accept POST requests from GitHub at the URL provided. Add this webhook URL in your GitHub settings area (at the URL provided) so that the daemon will receive updates from GitHub when your repository is updated. Once this is done, the daemon will automatically build and deploy any changes that are made to the deployed branch.

### Release Streams

The version of Inertia you are using can be seen in Inertia's `.inertia.toml` configuration file, or by running `inertia --version`. The version in `.inertia.toml` is used to determine what version of the Inertia daemon to use when you run `inertia $VPS_NAME init`.

You can manually change the daemon version used by editing the Inertia configuration file. If you are building from source, you can also check out the desired version and run `make inertia-tagged` or `make RELEASE=$STREAM`. Inertia daemon releases are tagged as follows:

- `v0.x.x` denotes [official, tagged releases](https://github.com/ubclaunchpad/inertia/releases) - these are recommended.
- `latest` denotes the newest builds on `master`.
- `canary` denotes experimental builds used for testing and development - do not use this.

The daemon component of an Inertia release can be patched separately from the CLI component - see our [wiki](https://github.com/ubclaunchpad/inertia/wiki/Daemon-Releases) for more details.

### Swag

Add some bling to your Inertia-deployed project :sunglasses: 

[![Deployed with Inertia](https://img.shields.io/badge/deploying%20with-inertia-blue.svg)](https://github.com/ubclaunchpad/inertia)

```
[![Deployed with Inertia](https://img.shields.io/badge/deploying%20with-inertia-blue.svg)](https://github.com/ubclaunchpad/inertia)
```

<br><br>

# :bulb: Motivation and Design

[UBC Launch Pad](http://www.ubclaunchpad.com) is a student-run software engineering club at the University of British Columbia that aims to provide students with a community where they can work together to build a all sorts of cool projects, ranging from mobile apps and web services to cryptocurrencies and machine learning applications.

Many of our projects rely on hosting providers for deployment. Unfortunately we frequently change hosting providers based on available funding and sponsorship, meaning our projects often need to be redeployed. On top of that, deployment itself can already be a frustrating task, especially for students with little to no experience setting up applications on remote hosts. Inertia is a project we started to address these problems, with the goal of developing an in-house deployment system that can make setting up continuously deployed applications simple and painless, regardless of the hosting provider.

The primary design goals of Inertia are to:

* minimize setup time for new projects
* maximimise compatibility across different client and VPS platforms
* offer an easy-to-learn interface for managing deployed applications

### How It Works

There is a detailed [Medium post](https://medium.com/ubc-launch-pad-software-engineering-blog/building-continuous-deployment-87a2bd8eedbe) that goes over the project, its motivations, the design choices we made, and Inertia's implementation. The team has also made a few presentations about Inertia that go over its design in some more detail:
- [UBC Launch Pad internal demo](https://slides.ubclaunchpad.com/projects/inertia/demo-1.pdf)
- [Vancouver DevOpsDays 2018](https://slides.ubclaunchpad.com/projects/inertia/devopsdays.pdf) ([video](https://youtu.be/amBYMEKGzTs?t=4h59m5s))

In summary, Inertia consists of two major components: a deployment daemon and a command line interface.

The deployment daemon runs persistently in the background on the server, receiving webhook events from GitHub whenever new commits are pushed. The CLI provides an interface to adjust settings and manage the deployment - this is done through HTTPS requests to the daemon, authenticated using JSON web tokens generated by the daemon. Remote configuration is stored locally in `.inertia.toml`.

<p align="center">
  <img src="https://bobheadxi.github.io/assets/images/posts/inertia-diagram.png" width="70%" />
</p>

Inertia is set up serverside by executing a script over SSH that installs Docker and starts an Inertia daemon image with [access to the host Docker socket](https://bobheadxi.github.io/dockerception/#docker-in-docker). This Docker-in-Docker configuration gives the daemon the ability to start up other containers *alongside* it, rather than *within* it, as required. Once the daemon is set up, we avoid using further SSH commands and execute Docker commands through Docker's Golang API. Instead of installing the docker-compose toolset, we [use a docker-compose image](https://bobheadxi.github.io/dockerception/#docker-compose-in-docker) to build and deploy user projects.

<br><br>

# :books: Contributing

Any contribution (pull requests, feedback, bug reports, ideas, etc.) is welcome! 

Please see our [contribution guide](https://github.com/ubclaunchpad/inertia/blob/master/CONTRIBUTING.md) for contribution guidelines as well as a detailed guide to help you get started with Inertia's codebase.

<br>

[![0](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/images/0)](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/links/0)
[![1](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/images/1)](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/links/1)
[![2](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/images/2)](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/links/2)
[![3](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/images/3)](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/links/3)
[![4](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/images/4)](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/links/4)
[![5](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/images/5)](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/links/5)
[![6](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/images/6)](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/links/6)
[![7](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/images/7)](https://sourcerer.io/fame/bobheadxi/ubclaunchpad/inertia/links/7)
