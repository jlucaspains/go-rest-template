package db

import (
	"goapi-template/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBInitSuccess(t *testing.T) {
	result, err := Init("sqlite", ":memory:", true)

	assert.Nil(t, err)
	db, _ := result.DB()

	assert.Nil(t, db.Ping())

	body := &models.Person{Name: "Test", Email: "email@email.com"}
	tx := result.Create(body)

	assert.Nil(t, tx.Error)
}

func TestDBInitBadConnectionString(t *testing.T) {
	_, err := Init("postgres", "Justbad", true)

	assert.Error(t, err)
}

func TestDBInitWithoutMigrate(t *testing.T) {
	result, err := Init("sqlite", ":memory:", false)

	assert.Nil(t, err)

	body := &models.Person{ID: 1, Name: "Test", Email: "email@email.com"}
	tx := result.FirstOrCreate(body, 1)

	assert.Error(t, tx.Error)
}
