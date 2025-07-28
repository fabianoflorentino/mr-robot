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

type DefaultProcessGateway struct {
	URL        string
	Name       string
	timeout    time.Duration
	httpClient *http.Client
}

// Default returns the processor default name.
func (p *DefaultProcessGateway) Default() string {
	return "default"
}

func NewDefaultProcessor(url string) DefaultProcessGateway {
	return DefaultProcessGateway{
		URL:        url,
		timeout:    defaultTimeout,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
}

// NewDefaultProcessorWithTimeout creates a new DefaultProcessGateway with custom timeout
func NewDefaultProcessorWithTimeout(url string, timeout time.Duration) DefaultProcessGateway {
	return DefaultProcessGateway{
		URL:        url,
		timeout:    timeout,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// Process requests the payment processor to process the payment. It returns
// a boolean indicating if the payment was processed successfully and an error
// if any occurred.
func (p *DefaultProcessGateway) Process(payment *domain.Payment) (bool, error) {
	if err := p.validatePayment(payment); err != nil {
		return false, err
	}

	req, err := p.createRequest(payment)
	if err != nil {
		return false, err
	}

	resp, err := p.sendRequest(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return p.isSuccessResponse(resp), nil
}

// validatePayment validates the payment object
func (p *DefaultProcessGateway) validatePayment(payment *domain.Payment) error {
	if payment == nil {
		return fmt.Errorf("payment cannot be nil")
	}
	return nil
}

// createRequest creates an HTTP request for the payment
func (p *DefaultProcessGateway) createRequest(payment *domain.Payment) (*http.Request, error) {
	payload, err := json.Marshal(payment)
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
func (p *DefaultProcessGateway) sendRequest(req *http.Request) (*http.Response, error) {
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
func (p *DefaultProcessGateway) ensureHTTPClient() error {
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
func (p *DefaultProcessGateway) isSuccessResponse(resp *http.Response) bool {
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
