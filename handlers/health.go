package handlers

import (
	"goapi-template/models"
	"net/http"
)

// GetHealth godoc
//
//	@Summary	Determines if the app is healthy
//	@Schemes
//	@Description	Returns HTTP 200 if the app is healthy and 400 if not
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	models.HealthResult
//	@Failure		400	{object}	models.HealthResult
//	@Router			/health [get]
func (h Handlers) GetHealth(w http.ResponseWriter, r *http.Request) {
	isDbHealthy := true
	localDB, err := h.DB.DB()
	if err != nil {
		isDbHealthy = false
	}

	isDbHealthy = localDB.Ping() == nil
	dbHealth := models.HealthResultItem{Name: "DB", Healthy: isDbHealthy}
	result := &models.HealthResult{Healthy: isDbHealthy, Dependencies: []models.HealthResultItem{dbHealth}}

	status := http.StatusOK
	if !isDbHealthy {
		status = http.StatusInternalServerError
	}

	h.JSON(w, status, result)
}
