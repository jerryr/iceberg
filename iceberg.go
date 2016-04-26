package main

import (
	"fmt"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"time"
	"strconv"
)
func main() {
	fmt.Println("Starting Iceberg...")
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	services := make(map[string] *Service)
	timer := time.NewTicker(time.Second * 5)
	updateServices(cli, services)
	go func() {
		for _ = range timer.C {
			updateServices(cli, services)
		}
	}()
	select {

	}
}

func updateServices(cli *client.Client, services map[string] *Service) {
	fmt.Println("Updating...")
	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		labels := c.Labels
		svcname := labels["com.docker.compose.service"]
		svc, ok := services[svcname]
		if ! ok {
			svc = NewService(svcname)
			services[svcname] = svc
		}
		var str string
		str, ok = labels["iceberg.minimum_count"]
		if ! ok {
			fmt.Println("No value given for minumum count!!")
		} else {
			var min int64
			min, err = strconv.ParseInt(str, 10, 64)
			if err != nil {
				fmt.Printf("Could not parse %v into a float\n", str)
			} else {
				svc.min = min
			}
		}
		str, ok = labels["iceberg.kill_probability"]
		if ! ok {
			fmt.Println("No value given for kill probability!!")
		} else {
			var kp float64
			kp, err = strconv.ParseFloat(str, 64)
			if err != nil {
				fmt.Printf("Could not parse %v into an int\n", str)
			} else {
				svc.killProb = kp
			}
		}
	}
}
