package core

import "errors"

var (
	ErrPaymentNotProcessed     = errors.New("payment can't be processed")
	ErrPaymentProcessingFailed = errors.New("payment processing failed")
)
