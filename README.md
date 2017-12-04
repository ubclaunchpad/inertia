# 👩‍🚀 Inertia

Inertia makes it easy to set up automated deployment for Dockerized
applications.

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
