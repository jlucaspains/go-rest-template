package handlers

import (
	"goapi-template/db"
	"goapi-template/models"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestPostPersonSuccess(t *testing.T) {
	db := &QuerierMock{
		InsertPersonResult: db.Person{
			ID:        1,
			Name:      "Demo Company",
			Email:     "demo@company.com",
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}
	r := setup(db)

	person := models.Person{
		Name:  "Demo Company",
		Email: "demo@company.com",
	}

	code, body, _, err := makeRequest[models.IdResult](r, "POST", "/person", person)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, code)
	assert.Equal(t, 1, body.ID)
}

func TestPostPersonMissingName(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	person := models.Person{
		Name:  "",
		Email: "",
	}

	code, result, _, err := makeRequest[models.ErrorResult](r, "POST", "/person", person)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "Name is required", result.Errors[0])
	assert.Equal(t, "Email is required", result.Errors[1])
}

func TestPostPersonDuplicate(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	person := models.Person{
		Name:  "Test",
		Email: "test@test.com",
	}

	code, result, _, err := makeRequest[models.ErrorResult](r, "POST", "/person", person)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusConflict, code)
	assert.Equal(t, "Record duplication detected", result.Errors[0])
}

func TestPutPersonSuccess(t *testing.T) {
	db := &QuerierMock{
		UpdatePersonResult: db.Person{
			ID:        1,
			Name:      "Test",
			Email:     "mail@company.com",
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}
	r := setup(db)

	person := models.Person{
		ID:    1,
		Name:  "Test 2",
		Email: "mail@company.com",
	}

	code, result, _, err := makeRequest[models.Person](r, "PUT", "/person/1", person)

	assert.Nil(t, err)
	assert.Equal(t, 200, code)
	assert.Equal(t, "Test 2", result.Name)
}

func TestPutPersonValidation(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	person := models.Person{
		ID:    1,
		Name:  "",
		Email: "mail@company.com",
	}

	code, result, _, err := makeRequest[models.ErrorResult](r, "PUT", "/person/1", person)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, result.Errors[0], "Name is required")
}

func TestPutPersonMissing(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	person := models.Person{
		ID:    10,
		Name:  "Test",
		Email: "mail@company.com",
	}

	code, _, _, err := makeRequest[models.ErrorResult](r, "PUT", "/person/10", person)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestPutPersonBadUrl(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	code, _, _, err := makeRequest[models.ErrorResult](r, "PUT", "/person/a", &models.Person{})

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestGetPersonSuccess(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	person := &models.Person{
		Name:  "Test",
		Email: "mail@company.com",
	}

	code, result, _, err := makeRequest[models.Person](r, "GET", "/person/1", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, result.ID, 1)
	assert.Equal(t, result.Name, person.Name)
	assert.Equal(t, result.Email, person.Email)
}

func TestGetPersonNotFound(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	code, _, _, err := makeRequest[models.Person](r, "GET", "/person/1", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestGetPersonBadUrl(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	code, _, _, err := makeRequest[models.Person](r, "GET", "/person/a", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestDeletePerson(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	code, _, _, err := makeRequest[models.Person](r, "DELETE", "/person/1", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, code)
}

func TestDeletePersonNotFound(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	code, _, _, err := makeRequest[models.Person](r, "DELETE", "/person/1", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestDeletePersonBadUrl(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	code, _, _, err := makeRequest[models.Person](r, "DELETE", "/person/a", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
}
