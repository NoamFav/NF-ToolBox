package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nf-software/nf-toolbox/internal/license"
)

// HandleActivate activates a license on a new device
func (s *Server) HandleActivate(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var req license.ActivationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if req.MachineFingerprint == "" {
		writeError(w, http.StatusBadRequest, "machine_fingerprint required")
		return
	}

	// Get subscription
	sub, err := s.db.GetSubscription(userID)
	if err != nil {
		writeError(w, http.StatusForbidden, "no active subscription")
		return
	}

	if sub.Status != "active" {
		writeError(w, http.StatusForbidden, "subscription not active")
		return
	}

	// Check device limit
	count, err := s.db.CountActiveDevices(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check devices")
		return
	}

	if count >= sub.MaxDevices {
		writeError(w, http.StatusForbidden, fmt.Sprintf("device limit reached (%d/%d)", count, sub.MaxDevices))
		return
	}

	// Record activation
	if err := s.db.CreateActivation(userID, "", req.MachineFingerprint, req.Nickname, req.Platform, req.Arch); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to record activation")
		return
	}

	// Get user's tools
	tools, err := s.db.GetUserTools(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get tools")
		return
	}

	// Build license payload
	payload := license.LicensePayload{
		UserID:             userID,
		Entitlements:       tools,
		MachineFingerprint: req.MachineFingerprint,
		IssuedAt:           time.Now(),
		ExpiresAt:          sub.CurrentPeriodEnd,
		Plan:               sub.Plan,
	}

	// Sign license
	signedLicense, err := license.Sign(payload, s.cfg.PrivateKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sign license")
		return
	}

	writeJSON(w, http.StatusOK, license.ActivationResponse{
		License: signedLicense,
	})
}

// HandleRenew renews a license for an existing device
func (s *Server) HandleRenew(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var req license.RenewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check if activation exists
	exists, err := s.db.ActivationExists(userID, req.MachineFingerprint)
	if err != nil || !exists {
		writeJSON(w, http.StatusOK, license.RenewResponse{
			Denied: true,
			Reason: "device_not_found",
		})
		return
	}

	// Get subscription
	sub, err := s.db.GetSubscription(userID)
	if err != nil {
		writeJSON(w, http.StatusOK, license.RenewResponse{
			Denied: true,
			Reason: "no_subscription",
		})
		return
	}

	if sub.Status != "active" {
		writeJSON(w, http.StatusOK, license.RenewResponse{
			Denied: true,
			Reason: "subscription_" + sub.Status,
		})
		return
	}

	// Update last seen
	s.db.UpdateLastSeen(userID, req.MachineFingerprint)

	// Get user's tools
	tools, _ := s.db.GetUserTools(userID)

	// Build license payload
	payload := license.LicensePayload{
		UserID:             userID,
		Entitlements:       tools,
		MachineFingerprint: req.MachineFingerprint,
		IssuedAt:           time.Now(),
		ExpiresAt:          sub.CurrentPeriodEnd,
		Plan:               sub.Plan,
	}

	// Sign license
	signedLicense, err := license.Sign(payload, s.cfg.PrivateKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sign license")
		return
	}

	writeJSON(w, http.StatusOK, license.RenewResponse{
		License: signedLicense,
	})
}

// HandleGetLicense returns current license (for sync)
func (s *Server) HandleGetLicense(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	fingerprint := r.URL.Query().Get("fingerprint")

	if fingerprint == "" {
		writeError(w, http.StatusBadRequest, "fingerprint query param required")
		return
	}

	// Get subscription
	sub, err := s.db.GetSubscription(userID)
	if err != nil {
		writeError(w, http.StatusForbidden, "no subscription")
		return
	}

	// Get tools
	tools, _ := s.db.GetUserTools(userID)

	// Build license
	payload := license.LicensePayload{
		UserID:             userID,
		Entitlements:       tools,
		MachineFingerprint: fingerprint,
		IssuedAt:           time.Now(),
		ExpiresAt:          sub.CurrentPeriodEnd,
		Plan:               sub.Plan,
	}

	signedLicense, err := license.Sign(payload, s.cfg.PrivateKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sign license")
		return
	}

	writeJSON(w, http.StatusOK, signedLicense)
}
