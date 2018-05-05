# Client

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/ubclaunchpad/inertia/client)

This package contains Inertia's clientside configuration and interface to remote Inertia daemons.

This package can be imported for use. For example:

```go
package main

import "github.com/ubclaunchpad/inertia/client"

func main() {
    // Set up Inertia
    config := client.NewConfig(
        "0.3.0", "inertia-deploy-test", "docker-compose",
    )

    // Add your remote
    config.AddRemote(&client.RemoteVPS{
		Name: "gcloud", // ...params
	})

    // Set up client, remote, and deploy your project
	cli, _ := client.NewClient("gcloud", config)
	cli.BootstrapRemote()
	cli.Up("git@github.com:ubclaunchpad/inertia.git", "", false)
}
```
