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