# Client

[![PkgGoDev](https://pkg.go.dev/badge/github.com/ubclaunchpad/inertia)](https://pkg.go.dev/github.com/ubclaunchpad/inertia/client)

This package contains Inertia's clientside interface to remote Inertia daemons. It can be imported for use if you don't like the CLI - for example:

```go
package main

import (
	"context"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/client"
	"github.com/ubclaunchpad/inertia/client/bootstrap"
)

func main() {
	// set up a client to your remote
	var inertia = client.NewClient(
		&cfg.Remote{
			Version: "v0.6.0",
			Name:    "gcloud",
			IP:      "my.host.addr",
			/* ... */
		},
		client.Options{ /* ... */ })

	// bootstrap your remote
	bootstrap.Bootstrap(inertia, bootstrap.Options{ /* ... */ })

	// deploy your project!
	inertia.Up(context.Background(), client.UpRequest{
		Project: "my-project",
		URL:     "git@github.com:me/project.git",
		Profile: cfg.Profile{
			Name:   "default",
			Branch: "master",
			Build: &cfg.Build{
				Type:          cfg.DockerCompose,
				BuildFilePath: "Dockerfile",
			},
		},
	})
}
```
