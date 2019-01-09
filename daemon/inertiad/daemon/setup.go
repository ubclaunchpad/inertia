package daemon

import (
	"context"
	"sync"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

func downloadDeps(cli *docker.Client, images ...string) {
	var wait sync.WaitGroup
	wait.Add(len(images))
	for _, i := range images {
		go dockerPull(i, cli, &wait)
	}
	wait.Wait()
	cli.Close()
}

func dockerPull(image string, cli *docker.Client, wait *sync.WaitGroup) {
	defer wait.Done()
	println("Downloading " + image)
	_, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		println(err.Error())
	} else {
		println(image + " download complete")
	}
}
