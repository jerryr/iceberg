package main

import (
	"fmt"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"strconv"
	"encoding/json"
	"github.com/docker/engine-api/types/events"
	"io"
	"github.com/docker/engine-api/types/filters"
)
const (
	ServiceNameLabel = "com.docker.compose.service"
	MinimumCountLabel string = "iceberg.minimum_count"
	KillProbabilityLabel string = "iceberg.kill_probability"
)

func main() {
	fmt.Println("Starting Iceberg...")
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	services := make(map[string] *Service)
	updateServices(cli, services)
	filter := filters.NewArgs()
	filter.Add("type", "container")
	filter.Add("event", "start")
	filter.Add("event", "die")
	read, err := cli.Events(context.Background(), types.EventsOptions{Filters: filter})
	if err != nil {
		panic("Cannot get event stream from Docker")
	}
	defer read.Close()
	dec := json.NewDecoder(read)
	for {
		var event events.Message
		err := dec.Decode(&event)
		if err != nil && err == io.EOF {
			break;
		}
		svcname := event.Actor.Attributes[ServiceNameLabel]
		evt := event.Action
		containerid := event.ID
		service, ok := services[svcname]
		if !ok {
			service = NewService(svcname)
			services[svcname] = service
		}
		updateVariables(service, event.Actor.Attributes)
		fmt.Println("Event type = ", evt)
		switch evt {
		case "start":
			service.AddContainer(containerid)
		case "die":
			service.RemoveContainer(containerid)
		}
		fmt.Printf("Service updated: %+v\n", service)
	}

	//timer := time.NewTicker(time.Second * 5)
	//go func() {
	//	for _ = range timer.C {
	//		updateServices(cli, services)
	//	}
	//}()
	select {
		// Sleep forever
	}
}

func updateVariables(svc *Service, labels map[string] string) {
	str, ok := labels[MinimumCountLabel]
	if ! ok {
		fmt.Println("No value given for minumum count!!")
	} else {
		var min int64
		min, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			fmt.Printf("Could not parse %v into a float\n", str)
		} else {
			svc.min = min
		}
	}
	str, ok = labels[KillProbabilityLabel]
	if ! ok {
		fmt.Println("No value given for kill probability!!")
	} else {
		var kp float64
		kp, err := strconv.ParseFloat(str, 64)
		if err != nil {
			fmt.Printf("Could not parse %v into an int\n", str)
		} else {
			svc.killProb = kp
		}
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
		svcname := labels[ServiceNameLabel]
		svc, ok := services[svcname]
		if ! ok {
			svc = NewService(svcname)
			services[svcname] = svc
		}
		updateVariables(svc, labels)
	}
}
