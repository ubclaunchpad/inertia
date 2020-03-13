/*

Inertia is the command line interface that helps you set up your remote for
continuous deployment and allows you to manage your deployment through
configuration options and various commands.

This document contains basic usage instructions, but a new usage guide is also
available here: https://inertia.ubclaunchpad.com/

Inertia can be installed in several ways:

	# Mac users
	brew install ubclaunchpad/tap/inertia

	# Windows users
	scoop bucket add ubclaunchpad https://github.com/ubclaunchpad/scoop-bucket
	scoop install inertia

Users of other platforms can install the Inertia CLI from the Releases page,
found here: https://github.com/ubclaunchpad/inertia/releases/latest

To help with usage, most relevant documentation can be seen by using the --help
flag on any command:

	inertia --help
	inertia init --help
	inertia [remote] up --help

Documentation can also be triggered by simply entering a command without the
prerequisite arguments or additional commands:

	inertia remote               # documentation about remote configuration

Inertia has two "core" sets of commands - one that primarily handles local
configuration, and one that allows you to control your remote VPS instances and
their associated deployments.

For local configuration, most commands will build off of the root "inertia ..."
command. For example, a typical set of commands to set up a project might look
like:

	inertia init                 # initiates Inertia configuration
	inertia remote add my_cloud  # adds configuration for a remote VPS instance

The other set of commands are based on a remote VPS configuration, and the
available commands can be seen by running:

	inertia [remote] --help

In the previous example, the next steps to set up a deployment might be:

	inertia my_cloud init        # bootstraps remote and installs Inertia daemon
	inertia my_cloud up          # deploys your project

Some of these commands offer a --stream flag that allows you to view realtime
log feedback from the daemon.

More documentation on Inertia, how it works, and how to use it can be found
in the project repository: https://github.com/ubclaunchpad/inertia/tree/master

*/
package main
