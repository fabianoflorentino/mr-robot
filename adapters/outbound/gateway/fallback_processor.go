package gateway

import "github.com/fabianoflorentino/mr-robot/core/domain"

type ProcessorFallbackClient struct{}

func (p *ProcessorFallbackClient) Fallback() string {
	return "fallback"
}

func (p *ProcessorFallbackClient) Process(payment *domain.Payment) (bool, error) {
	return true, nil
}
