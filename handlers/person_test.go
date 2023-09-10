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

func TestPostPersonSuccess(t *testing.T) {
	r, db := setup(true, true)

	handlers := &Handlers{DB: db}
	r.HandleFunc("/person", handlers.PostPerson).Methods("POST")

	person := models.Person{
		Name:  "Demo Company",
		Email: "demo@company.com",
	}

	code, body, err := makeRequest[models.IdResult](r, "POST", "/person", person)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, code)
	assert.Equal(t, 1, body.ID)
}

func TestPostPersonMissingName(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person", handlers.PostPerson).Methods("POST")

	person := models.Person{
		Name:  "",
		Email: "",
	}

	jsonValue, _ := json.Marshal(person)
	reqFound, _ := http.NewRequest("POST", "/person", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	result := &models.ErrorResult{}
	json.Unmarshal(w.Body.Bytes(), result)

	assert.Equal(t, "Name is required", result.Errors[0])
	assert.Equal(t, "Email is required", result.Errors[1])
}

func TestPostPersonDuplicate(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}

	r.HandleFunc("/person", handlers.PostPerson).Methods("POST")

	person := models.Person{
		Name:  "Test",
		Email: "test@test.com",
	}

	jsonValue, _ := json.Marshal(person)
	reqFound, _ := http.NewRequest("POST", "/person", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	jsonValue2, _ := json.Marshal(person)
	reqFound2, _ := http.NewRequest("POST", "/person", bytes.NewBuffer(jsonValue2))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, reqFound2)

	assert.Equal(t, http.StatusConflict, w2.Code)

	result := &models.ErrorResult{}
	json.Unmarshal(w2.Body.Bytes(), result)

	assert.Equal(t, "Record duplication detected", result.Errors[0])
}

func TestPutPersonSuccess(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.PutPerson).Methods("PUT")

	person := models.Person{
		ID:    1,
		Name:  "Test",
		Email: "mail@company.com",
	}

	db.Create(&person)

	person.Name = "Test 2"

	jsonValue, _ := json.Marshal(person)
	reqFound, _ := http.NewRequest("PUT", "/person/1", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	assert.Equal(t, http.StatusAccepted, w.Code)

	result := &models.Person{}
	db.Find(result, 1)

	assert.Equal(t, "Test 2", result.Name)
}

func TestPutPersonValidation(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.PutPerson).Methods("PUT")

	person := models.Person{
		ID:    1,
		Name:  "Test",
		Email: "mail@company.com",
	}

	db.Create(&person)

	person.Name = ""

	jsonValue, _ := json.Marshal(person)
	reqFound, _ := http.NewRequest("PUT", "/person/1", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	result := &models.ErrorResult{}
	json.Unmarshal(w.Body.Bytes(), result)

	assert.Equal(t, result.Errors[0], "Name is required")
}

func TestPutPersonMissing(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.PutPerson).Methods("PUT")

	person := models.Person{
		ID:    10,
		Name:  "Test",
		Email: "mail@company.com",
	}

	jsonValue, _ := json.Marshal(person)
	reqFound, _ := http.NewRequest("PUT", "/person/10", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPutPersonBadUrl(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.PutPerson).Methods("PUT")

	reqFound, _ := http.NewRequest("PUT", "/person/a", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPersonSuccess(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.GetPerson).Methods("GET")

	person := &models.Person{
		Name:  "Test",
		Email: "mail@company.com",
	}
	db.Create(person)

	jsonValue, _ := json.Marshal(person)
	reqFound, _ := http.NewRequest("GET", "/person/1", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	assert.Equal(t, http.StatusOK, w.Code)

	result := &models.Person{}
	json.Unmarshal(w.Body.Bytes(), result)

	assert.Equal(t, result.ID, 1)
	assert.Equal(t, result.Name, person.Name)
	assert.Equal(t, result.Email, person.Email)
}

func TestGetPersonNotFound(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.GetPerson).Methods("GET")

	reqFound, _ := http.NewRequest("GET", "/person/1", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetPersonBadUrl(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.GetPerson).Methods("GET")

	reqFound, _ := http.NewRequest("GET", "/person/a", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePerson(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.DeletePerson).Methods("DELETE")

	person := &models.Person{
		Name:  "Test",
		Email: "mail@company.com",
	}
	db.Create(person)

	reqFound, _ := http.NewRequest("DELETE", "/person/1", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	verification := &models.Person{}
	db.First(verification, 1)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Equal(t, 0, verification.ID)
}

func TestDeletePersonNotFound(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.DeletePerson).Methods("DELETE")

	reqFound, _ := http.NewRequest("DELETE", "/person/1", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeletePersonBadUrl(t *testing.T) {
	r, db := setup(true, true)
	handlers := &Handlers{DB: db}
	r.HandleFunc("/person/{id}", handlers.DeletePerson).Methods("DELETE")

	reqFound, _ := http.NewRequest("DELETE", "/person/a", bytes.NewBuffer([]byte{}))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, reqFound)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
