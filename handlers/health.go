package handlers

import (
	"goapi-template/models"
	"net/http"

	"github.com/gin-gonic/gin"
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
func (h Handlers) GetHealth(c *gin.Context) {
	isDbHealthy := true
	localDB, err := h.DB.DB()
	if err != nil {
		isDbHealthy = false
	}

	isDbHealthy = localDB.Ping() == nil
	dbHealth := models.HealthResultItem{Name: "DB", Healthy: isDbHealthy}

	if isDbHealthy {
		c.JSON(http.StatusOK, &models.HealthResult{Healthy: isDbHealthy, Dependencies: []models.HealthResultItem{dbHealth}})
	} else {
		c.JSON(http.StatusInternalServerError, &models.HealthResult{Healthy: isDbHealthy, Dependencies: []models.HealthResultItem{dbHealth}})
	}
}
