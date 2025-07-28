package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fabianoflorentino/mr-robot/core/domain"
)

type DefaultProcessGateway struct {
	URL  string
	Name string
}

// Default returns the processor default name.
func (p *DefaultProcessGateway) Default() string {
	return "default"
}

func NewDefaultProcessor(url string) DefaultProcessGateway {
	return DefaultProcessGateway{URL: url}
}

// Process requests the payment processor to process the payment. It returns
// a boolean indicating if the payment was processed successfully and an error
// if any occurred.
func (p *DefaultProcessGateway) Process(payment *domain.Payment) (bool, error) {
	payload, err := json.Marshal(payment)
	if err != nil {
		return false, fmt.Errorf("error to serialize payment: %w", err)
	}

	req, err := http.NewRequest("POST", p.URL, bytes.NewBuffer(payload))
	if err != nil {
		return false, fmt.Errorf("error to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send payment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, nil
	}

	return false, nil
}
