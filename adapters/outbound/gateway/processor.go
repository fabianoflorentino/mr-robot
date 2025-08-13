package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fabianoflorentino/mr-robot/core/domain"
)

const (
	defaultTimeout  = 5 * time.Second
	contentType     = "Content-Type"
	applicationJson = "application/json"
	httpMethodPost  = "POST"
)

type ProcessGateway struct {
	URL        string
	Name       string
	timeout    time.Duration
	httpClient *http.Client
}

func NewProcessor(p *ProcessGateway) ProcessGateway {
	return ProcessGateway{
		URL:        p.URL,
		timeout:    p.timeout,
		httpClient: p.httpClient,
	}
}

// ProcessorName returns the processor name.
func (p *ProcessGateway) ProcessorName() string {
	if p.Name == "" {
		return "default"
	}
	return p.Name
}

// Process requests the payment processor to process the payment. It returns
// a boolean indicating if the payment was processed successfully and an error
// if any occurred.
func (p *ProcessGateway) Process(payment *domain.Payment) (bool, error) {
	if err := p.validatePayment(payment); err != nil {
		return false, err
	}

	req, err := p.createRequest(payment)
	if err != nil {
		return false, err
	}

	resp, err := p.sendRequest(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request to %s: %w", p.ProcessorName(), err)
	}
	defer resp.Body.Close()

	success := p.isSuccessResponse(resp)
	if !success {
		return false, fmt.Errorf("payment processing failed: HTTP %d from %s", resp.StatusCode, p.ProcessorName())
	}

	return success, nil
}

// validatePayment validates the payment object
func (p *ProcessGateway) validatePayment(payment *domain.Payment) error {
	if payment == nil {
		return fmt.Errorf("payment cannot be nil")
	}
	return nil
}

// createRequest creates an HTTP request for the payment
func (p *ProcessGateway) createRequest(payment *domain.Payment) (*http.Request, error) {
	processorPayment := map[string]any{"correlationId": payment.CorrelationID, "amount": payment.Amount}

	payload, err := json.Marshal(processorPayment)
	if err != nil {
		return nil, fmt.Errorf("error to serialize payment: %w", err)
	}

	req, err := http.NewRequest(httpMethodPost, p.URL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("error to create request: %w", err)
	}

	req.Header.Set(contentType, applicationJson)
	return req, nil
}

// sendRequest sends the HTTP request using the configured client
func (p *ProcessGateway) sendRequest(req *http.Request) (*http.Response, error) {
	if err := p.ensureHTTPClient(); err != nil {
		return nil, ErrHttpClientNotInitialized
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send payment: %w", err)
	}
	return resp, nil
}

// ensureHTTPClient validates and initializes the HTTP client if needed
func (p *ProcessGateway) ensureHTTPClient() error {
	if p.httpClient == nil {
		timeout := defaultTimeout

		if p.timeout > 0 {
			timeout = p.timeout
		}

		p.httpClient = &http.Client{Timeout: timeout}
	}

	return nil
}

// isSuccessResponse checks if the response indicates success
func (p *ProcessGateway) isSuccessResponse(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
