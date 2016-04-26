package main

import (
	"fmt"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"time"
)
func main() {
	fmt.Println("Starting Iceberg...")
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	services := make(map[string] Service)
	timer := time.NewTicker(time.Second * 5)
	go func() {
		for _ = range timer.C {
			updateServices(cli, services)
		}
	}()
	select {
	
	}
}

func updateServices(cli *client.Client, services map[string] Service) {
	fmt.Println("Updating...")
	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		fmt.Println(c.ID)
	}
}
