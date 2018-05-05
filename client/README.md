# Client

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/ubclaunchpad/inertia/client)

This package contains Inertia's clientside configuration and interface to remote Inertia daemons.

This package can be imported for use. For example:

```go
package main

import "github.com/ubclaunchpad/inertia/client"

func main() {
    remote := &client.RemoteVPS{
		Name: "gcloud",
		/* ... */
	}
	config := &client.Config{
		Version:   "0.3.0",
		Project:   "inertia-deploy-test",
		BuildType: "docker-compose",
		Remotes:   []*client.RemoteVPS{remote},
	}

	client, _ := config.NewClient("gcloud")
	client.BootstrapRemote()
	client.Up("git@github.com:ubclaunchpad/inertia.git", "", false)
}
```
