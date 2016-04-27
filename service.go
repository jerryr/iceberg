package main

type Service struct {
	id string
	min int64
	killProb float64
	running map[string] bool
}

func NewService(svcname string) *Service {
	return &Service{id: svcname, running: make(map[string] bool)}
}

func (svc *Service) AddContainer(containerid string) {
	svc.running[containerid] = true
}

func (svc *Service) RemoveContainer(containerid string) {
	delete(svc.running, containerid)
}