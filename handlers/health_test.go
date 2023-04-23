package handlers

import (
	"bytes"
	"encoding/json"
	"goapi-template/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthSuccess(t *testing.T) {
	r, db := setup(true, false)
	handlers := &Handlers{DB: db}
	r.GET("/health", handlers.GetHealth)

	reqFound, _ := http.NewRequest("GET", "/health", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	result := &models.HealthResult{}
	json.Unmarshal(w.Body.Bytes(), result)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, result.Healthy)
	assert.True(t, result.Dependencies[0].Healthy)
}

func TestHealthBadDB(t *testing.T) {
	r, db := setup(true, false)
	handlers := &Handlers{DB: db}
	r.GET("/health", handlers.GetHealth)

	newDB, _ := db.DB()
	newDB.Close()

	reqFound, _ := http.NewRequest("GET", "/health", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	result := &models.HealthResult{}
	json.Unmarshal(w.Body.Bytes(), result)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.False(t, result.Healthy)
	assert.False(t, result.Dependencies[0].Healthy)
}
