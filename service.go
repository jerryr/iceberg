package main

import (
	"math/rand"
	"github.com/docker/distribution/context"
	"os/exec"
	"strconv"
)

type Service struct {
	id string
	min int
	killProb float64
	running map[string] bool
	chaosActive bool
}

func NewService(svcname string) *Service {
	svc := &Service{id: svcname, running: make(map[string] bool)}
	return svc
}

func (svc *Service) AddContainer(containerid string) {
	svc.running[containerid] = true
}

func (svc *Service) RemoveContainer(containerid string) {
	delete(svc.running, containerid)
}

func (svc *Service) chaosify() {
	println("Chaosing", svc.id)
	if rand.Float64() <= svc.killProb && len(svc.running) > svc.min {
		cankill := len(svc.running) - svc.min
		tokill := rand.Intn(cankill) + 1
		println(svc.id, "tokill =", tokill)
		// first collect the container ids in a slice
		ids := []string{}
		for k := range svc.running {
			ids = append(ids, k)
		}
		kill := []string{}
		for n := 0; n < tokill; n++ {
			idx := rand.Intn(len(ids))
			kill = append(kill, ids[idx])
			ids = append(ids[:idx], ids[idx+1:]...)
		}
		for _, container := range kill {
			go killContainer(container)
		}

	}
}

func killContainer(containerid string) {
	println ("Killing ", containerid)
	data, err := dockerClient.ContainerInspect(context.Background(), containerid)
	if err != nil {
		println ("Could not inspect container", containerid)
		return
	}
	pid := data.State.Pid
	cmd := exec.Command("docker-machine", "ssh", "default", "sudo", "kill", "-9", strconv.Itoa(pid))
	err = cmd.Run()
	if err != nil {
		println ("Could not kill container ", containerid, " err = ", err)
	}
}