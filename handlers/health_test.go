package handlers

import (
	"errors"
	"goapi-template/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthSuccess(t *testing.T) {
	db := &QuerierMock{PingDbResult: 1}
	r := setup(db)

	code, result, _, err := makeRequest[models.HealthResult](r, "GET", "/health", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, result.Healthy)
	assert.True(t, result.Dependencies[0].Healthy)
}

func TestHealthBadDB(t *testing.T) {
	db := &QuerierMock{PingDbError: errors.New("Bad DB")}
	r := setup(db)

	code, result, _, err := makeRequest[models.HealthResult](r, "GET", "/health", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.False(t, result.Healthy)
	assert.False(t, result.Dependencies[0].Healthy)
}
