package main

import (
	"encoding/json"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/events"
	"github.com/docker/engine-api/types/filters"
	"golang.org/x/net/context"
	"io"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	ServiceNameLabel     string = "com.docker.compose.service"
	MinimumCountLabel    string = "iceberg.minimum_count"
	KillProbabilityLabel string = "iceberg.kill_probability"
	AutostartLabel       string = "iceberg.autostart"
)

var rwlock sync.RWMutex
var dockerClient *client.Client

func main() {
	log.Println("Starting Iceberg...")
	ticker := time.NewTicker(time.Second * 5)
	rand.Seed(time.Now().Unix())
	cli, err := client.NewEnvClient()
	dockerClient = cli
	if err != nil {
		panic(err)
	}
	services := make(map[string]*Service)
	updateServices(cli, services)

	go func() {
		for _ = range ticker.C {
			performChaos(services)
		}
	}()
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
	log.Println("Listening for Docker events...")
	for {
		var event events.Message
		err := dec.Decode(&event)
		if err != nil && err == io.EOF {
			break
		}
		svcname := event.Actor.Attributes[ServiceNameLabel]
		evt := event.Action
		containerid := event.ID
		service, ok := services[svcname]
		if !ok {
			log.Println("Detected a new service", svcname)
			service = NewService(svcname)
			rwlock.Lock()
			services[svcname] = service
			rwlock.Unlock()
		}
		updateVariables(service, event.Actor.Attributes)
		//fmt.Println("Event type = ", evt)
		switch evt {
		case "start":
			service.AddContainer(containerid)
		case "die":
			service.RemoveContainer(containerid)
		}
		//fmt.Printf("Service updated: %+v\n", service)
	}

	select {
	// Sleep forever
	}
}

func updateVariables(svc *Service, labels map[string]string) {
	rwlock.Lock()
	str, ok := labels[MinimumCountLabel]
	if ok {
		var min int
		min, err := strconv.Atoi(str)
		if err != nil {
			log.Printf("Could not parse %v into a float\n", str)
		} else {
			svc.min = min
		}
	}
	str, ok = labels[KillProbabilityLabel]
	if ok {
		var kp float64
		kp, err := strconv.ParseFloat(str, 64)
		if err != nil {
			log.Printf("Could not parse %v into an int\n", str)
		} else {
			svc.killProb = kp
		}
	}

	str, ok = labels[AutostartLabel]
	if ok {
		var as bool
		as, err := strconv.ParseBool(str)
		if err == nil {
			svc.chaosActive = as
		}
	}
	rwlock.Unlock()

}

func updateServices(cli *client.Client, services map[string]*Service) {
	log.Println("Getting initial state...")
	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}
	// Not locking because the eventing stuff isnt active yet
	//rwlock.Lock()

	for _, c := range containers {
		labels := c.Labels
		svcname := labels[ServiceNameLabel]
		svc, ok := services[svcname]
		if !ok {
			log.Println("Detected a new service", svcname)
			svc = NewService(svcname)
			services[svcname] = svc
		}
		updateVariables(svc, labels)
		if c.State == "running" {
			svc.AddContainer(c.ID)
		}
	}

	//rwlock.Unlock()
}

func performChaos(services map[string]*Service) {
	rwlock.RLock()
	for _, svc := range services {
		if svc.chaosActive {
			svc.chaosify()
		}
	}
	rwlock.RUnlock()
}
