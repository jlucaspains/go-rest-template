package handlers

import (
	"fmt"
	"goapi-template/db"
	"goapi-template/models"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	db := &QuerierMock{
		InsertPersonError: &pgconn.PgError{Code: "23505"},
	}
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
		UpdatePersonResult: 1,
	}
	r := setup(db)

	person := models.Person{
		ID:    1,
		Name:  "Test 2",
		Email: "mail@company.com",
	}

	code, _, _, err := makeRequest[string](r, "PUT", "/person/1", person)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, code)
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

func TestPutPersonDbError(t *testing.T) {
	db := &QuerierMock{
		UpdatePersonError: fmt.Errorf("db error"),
	}
	r := setup(db)

	person := models.Person{
		ID:    1,
		Name:  "test",
		Email: "mail@company.com",
	}

	code, result, _, err := makeRequest[models.ErrorResult](r, "PUT", "/person/1", person)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, result.Errors[0], "Unknown error")
}

func TestPutPersonMissing(t *testing.T) {
	db := &QuerierMock{
		UpdatePersonResult: 0,
	}
	r := setup(db)

	person := models.Person{
		ID:    10,
		Name:  "Test",
		Email: "mail@company.com",
	}

	code, _, _, err := makeRequest[string](r, "PUT", "/person/10", person)

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
	db := &QuerierMock{
		GetPersonByIdResult: db.Person{
			ID:        1,
			Name:      "Test",
			Email:     "mail@company.com",
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}
	r := setup(db)

	person := &models.Person{
		Name:  "Test",
		Email: "mail@company.com",
	}

	code, result, _, err := makeRequest[models.Person](r, "GET", "/person/1", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, person.Name, result.Name)
	assert.Equal(t, person.Email, result.Email)
}

func TestGetPersonNotFound(t *testing.T) {
	db := &QuerierMock{
		GetPersonByIdError: pgx.ErrNoRows,
	}
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
	db := &QuerierMock{
		DeletePersonResult: 1,
	}
	r := setup(db)

	code, _, _, err := makeRequest[string](r, "DELETE", "/person/1", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, code)
}

func TestDeletePersonNotFound(t *testing.T) {
	db := &QuerierMock{
		DeletePersonResult: 0,
	}
	r := setup(db)

	code, _, _, err := makeRequest[string](r, "DELETE", "/person/1", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestDeletePersonBadUrl(t *testing.T) {
	db := &QuerierMock{}
	r := setup(db)

	code, _, _, err := makeRequest[string](r, "DELETE", "/person/a", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestDeletePersonDbError(t *testing.T) {
	db := &QuerierMock{
		DeletePersonError: fmt.Errorf("db error"),
	}
	r := setup(db)

	code, result, _, err := makeRequest[models.ErrorResult](r, "DELETE", "/person/1", nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, result.Errors[0], "Unknown error")
}
