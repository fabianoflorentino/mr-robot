package client

import "github.com/fabianoflorentino/mr-robot/core/domain"

type ProcessorDefaultClient struct{}

func (p *ProcessorDefaultClient) Default() string {
	return "default"
}

func (p *ProcessorDefaultClient) Process(payment *domain.Payment) (bool, error) {
	return true, nil
}
