/*

Inertiad is Inertia's daemon component.

This service runs in the background on your remote VPS and allows you to monitor
and control your deployed application.

Even though it is built as command line application, inertiad not intended for
direct use - the Inertia daemon is supposed to be deployed as a Docker container,
the image for which can be found here:
https://hub.docker.com/r/ubclaunchpad/inertia/

When used, however, it offers two main commands:

	inertiad token   # generates and outputs a JWT in stdout
	inertiad run     # starts daemon service

*/
package main
