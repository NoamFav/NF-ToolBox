package api

import (
	"net/http"
)

// HandleListTools returns catalog of all tools with ownership status
func (s *Server) HandleListTools(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	// Get all tools
	allTools, err := s.db.GetAllTools()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get tools")
		return
	}

	// Get user's tools
	userTools, _ := s.db.GetUserTools(userID)
	ownedMap := make(map[string]bool)
	for _, t := range userTools {
		ownedMap[t] = true
	}

	// Build response
	type ToolResponse struct {
		Name         string `json:"name"`
		DisplayName  string `json:"display_name"`
		Description  string `json:"description"`
		PriceMonthly int    `json:"price_monthly"`
		PriceYearly  int    `json:"price_yearly"`
		Owned        bool   `json:"owned"`
	}

	var tools []ToolResponse
	for _, t := range allTools {
		tools = append(tools, ToolResponse{
			Name:         t.Name,
			DisplayName:  t.DisplayName,
			Description:  t.Description,
			PriceMonthly: t.PriceMonthly,
			PriceYearly:  t.PriceYearly,
			Owned:        ownedMap[t.Name],
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tools": tools,
	})
}
