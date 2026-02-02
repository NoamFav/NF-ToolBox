package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/nf-software/nf-toolbox/internal/db"
)

// HandleStripeWebhook processes Stripe webhook events
// TODO: Add proper Stripe signature verification
func (s *Server) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	// TODO: Verify Stripe signature
	// sig := r.Header.Get("Stripe-Signature")
	// event, err := webhook.ConstructEvent(body, sig, s.cfg.StripeWebhookSecret)

	var event struct {
		Type string `json:"type"`
		Data struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	switch event.Type {
	case "customer.subscription.created", "customer.subscription.updated":
		var sub struct {
			ID                   string `json:"id"`
			Customer             string `json:"customer"`
			Status               string `json:"status"`
			CurrentPeriodEnd     int64  `json:"current_period_end"`
			Metadata             map[string]string `json:"metadata"`
			Items                struct {
				Data []struct {
					Plan struct {
						Interval string `json:"interval"`
					} `json:"plan"`
				} `json:"data"`
			} `json:"items"`
		}
		if err := json.Unmarshal(event.Data.Object, &sub); err != nil {
			writeError(w, http.StatusBadRequest, "invalid subscription data")
			return
		}

		userID := sub.Metadata["user_id"]
		if userID == "" {
			writeError(w, http.StatusBadRequest, "missing user_id in metadata")
			return
		}

		plan := "monthly"
		if len(sub.Items.Data) > 0 && sub.Items.Data[0].Plan.Interval == "year" {
			plan = "yearly"
		}

		dbSub := &db.Subscription{
			UserID:           userID,
			Status:           sub.Status,
			Plan:             plan,
			CurrentPeriodEnd: time.Unix(sub.CurrentPeriodEnd, 0),
			MaxDevices:       3,
			StripeCustomerID: sub.Customer,
			StripeSubID:      sub.ID,
		}

		if err := s.db.UpsertSubscription(dbSub); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update subscription")
			return
		}

	case "customer.subscription.deleted":
		var sub struct {
			ID       string `json:"id"`
			Metadata map[string]string `json:"metadata"`
		}
		if err := json.Unmarshal(event.Data.Object, &sub); err != nil {
			writeError(w, http.StatusBadRequest, "invalid subscription data")
			return
		}

		userID := sub.Metadata["user_id"]
		if userID == "" {
			break
		}

		// Mark as canceled
		existing, err := s.db.GetSubscription(userID)
		if err == nil {
			existing.Status = "canceled"
			s.db.UpsertSubscription(existing)
		}
	}

	w.WriteHeader(http.StatusOK)
}
