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
	URL string
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
func (p *DefaultProcessGateway) DefaultProcess(payment *domain.Payment) (bool, error) {
	req, err := http.NewRequest("POST", p.URL, bytes.NewBuffer(paymentMarshal(payment)))
	if err != nil {
		return false, fmt.Errorf("error to request the process: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error: cant reach the server: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}

// paymentMarshal marshals the given payment into a JSON byte slice.
func paymentMarshal(payment *domain.Payment) []byte {
	var b bytes.Buffer

	e := json.NewEncoder(&b)
	e.Encode(payment)

	return b.Bytes()
}
