package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// HandleListDevices returns all active devices for user
func (s *Server) HandleListDevices(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	activations, err := s.db.GetActivations(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get devices")
		return
	}

	// Get subscription for max devices
	sub, _ := s.db.GetSubscription(userID)
	maxDevices := 3
	if sub != nil {
		maxDevices = sub.MaxDevices
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"devices":     activations,
		"max_devices": maxDevices,
		"count":       len(activations),
	})
}

// HandleDeactivateDevice removes a device activation
func (s *Server) HandleDeactivateDevice(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	activationID := chi.URLParam(r, "id")

	if activationID == "" {
		writeError(w, http.StatusBadRequest, "device id required")
		return
	}

	err := s.db.DeactivateDevice(userID, activationID)
	if err != nil {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
