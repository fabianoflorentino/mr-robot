package domain

type Payment struct {
	CorrelationID string  `json:"correlation_id,omitempty"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
}
