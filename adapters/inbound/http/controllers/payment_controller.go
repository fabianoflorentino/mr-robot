package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fabianoflorentino/mr-robot/core"
	"github.com/fabianoflorentino/mr-robot/core/domain"
	"github.com/fabianoflorentino/mr-robot/internal/app/interfaces"
	"github.com/fabianoflorentino/mr-robot/internal/app/queue"
)

type PaymentController struct {
	q *queue.PaymentQueue
	s interfaces.PaymentServiceInterface
}

func NewPaymentController(q *queue.PaymentQueue, s interfaces.PaymentServiceInterface) *PaymentController {
	return &PaymentController{q: q, s: s}
}

func (u *PaymentController) PaymentProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var payment = &domain.Payment{}

	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "correlationId and amount are required")
		return
	}

	u.enqueuePaymentWithTimeout(w, payment)
}

func (u *PaymentController) PaymentsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var from, to *time.Time

	queryFrom := r.URL.Query().Get("from")
	queryTo := r.URL.Query().Get("to")

	// Parse query parameters for date range
	// If both are provided, parse them and set the from and to variables
	if queryFrom != "" && queryTo != "" {
		fromParsed, err := time.Parse(time.RFC3339, queryFrom)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "invalid from date format, use RFC3339 format, Ex: 2023-01-01T00:00:00Z")
			return
		}

		toParsed, err := time.Parse(time.RFC3339, queryTo)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "invalid to date format, use RFC3339 format, Ex: 2023-01-01T00:00:00Z")
			return
		}

		from = &fromParsed
		to = &toParsed

		// If only one of the dates is provided, return an error
	} else if queryFrom != "" || queryTo != "" {
		writeErrorResponse(w, http.StatusBadRequest, "both from and to dates must be provided")
		return
	}

	summary, err := u.s.Summary(r.Context(), from, to)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to retrieve payment summary", err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, summary)
}

func (u *PaymentController) PurgePayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := u.s.Purge(r.Context()); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to purge payments", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (u *PaymentController) enqueuePaymentWithTimeout(w http.ResponseWriter, payment *domain.Payment) {
	eq := make(chan error, 1)

	go func() { eq <- u.q.Enqueue(payment) }()

	select {
	case err := <-eq:
		if err != nil {
			if err == core.ErrQueueFull {
				writeErrorResponse(w, http.StatusTooManyRequests, "system is busy, please try again later")
				return
			}

			writeErrorResponse(w, http.StatusInternalServerError, "failed to process payment", err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)

	case <-time.After(TIME_AFTER):
		writeErrorResponse(w, http.StatusRequestTimeout, "request timeout", "unable to queue payment within timeout")
	}
}
