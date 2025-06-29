package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CacherMock struct {
	GetStringResult string
	GetStringError  error
	SetStringKey    string
	SetStringValue  string
	SetStringError  error
	DeleteKeyError  error
	DeleteKeyKey    string
}

func (m *CacherMock) GetString(ctx context.Context, key string) (string, error) {
	return m.GetStringResult, m.GetStringError
}

func (m *CacherMock) SetString(ctx context.Context, key string, value string) error {
	m.SetStringKey = key
	m.SetStringValue = value
	return m.SetStringError
}

func (m *CacherMock) DeleteKey(ctx context.Context, key string) error {
	m.DeleteKeyKey = key
	return m.DeleteKeyError
}

type QuerierMock struct {
	GetPeopleResult     []Person
	GetPeopleError      error
	GetPersonByIdResult Person
	GetPersonByIdError  error
	InsertPersonResult  Person
	InsertPersonError   error
	UpdatePersonResult  int64
	UpdatePersonError   error
	DeletePersonResult  int64
	DeletePersonError   error
	PingDbResult        int32
	PingDbError         error
}

func (m *QuerierMock) GetPeople(ctx context.Context) ([]Person, error) {
	return m.GetPeopleResult, m.GetPeopleError
}

func (m *QuerierMock) GetPersonById(ctx context.Context, id int32) (Person, error) {
	return m.GetPersonByIdResult, m.GetPersonByIdError
}

func (m *QuerierMock) InsertPerson(ctx context.Context, arg InsertPersonParams) (Person, error) {
	return m.InsertPersonResult, m.InsertPersonError
}

func (m *QuerierMock) UpdatePerson(ctx context.Context, arg UpdatePersonParams) (int64, error) {
	return m.UpdatePersonResult, m.UpdatePersonError
}

func (m *QuerierMock) DeletePerson(ctx context.Context, id int32) (int64, error) {
	return m.DeletePersonResult, m.DeletePersonError
}

func (m *QuerierMock) PingDb(ctx context.Context) (int32, error) {
	return m.PingDbResult, m.PingDbError
}

func TestNewCachingQuerier(t *testing.T) {
	result := NewCachingQuerier(nil, nil)

	assert.NotNil(t, result)
}

func TestGetPeopleWithCacheSuccess(t *testing.T) {
	querier := NewCachingQuerier(&QuerierMock{}, &CacherMock{
		GetStringResult: `[{"ID":1,"Name":"Test"}]`,
	})
	result, err := querier.GetPeople(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, int32(1), result[0].ID)
}

func TestGetPeopleWithCacheMiss(t *testing.T) {
	querier := NewCachingQuerier(&QuerierMock{
		GetPeopleResult: []Person{{ID: 1, Name: "Test"}},
	}, &CacherMock{})
	result, err := querier.GetPeople(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, int32(1), result[0].ID)
}

func TestGetPeopleWithCacheFail(t *testing.T) {
	querier := NewCachingQuerier(&QuerierMock{
		GetPeopleResult: []Person{{ID: 1, Name: "Test"}},
	}, &CacherMock{
		GetStringError: fmt.Errorf("error"),
	})
	result, err := querier.GetPeople(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, int32(1), result[0].ID)
}

func TestGetPersonWithCacheSuccess(t *testing.T) {
	querier := NewCachingQuerier(&QuerierMock{}, &CacherMock{
		GetStringResult: `{"ID":1,"Name":"Test"}`,
	})
	result, err := querier.GetPersonById(context.Background(), 1)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.ID)
}

func TestGetPersonWithCacheMiss(t *testing.T) {
	querier := NewCachingQuerier(&QuerierMock{
		GetPersonByIdResult: Person{ID: 1, Name: "Test"},
	}, &CacherMock{})
	result, err := querier.GetPersonById(context.Background(), 1)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.ID)
}

func TestGetPersonWithCacheFail(t *testing.T) {
	querier := NewCachingQuerier(&QuerierMock{
		GetPersonByIdResult: Person{ID: 1, Name: "Test"},
	}, &CacherMock{
		GetStringError: fmt.Errorf("error"),
	})
	result, err := querier.GetPersonById(context.Background(), 1)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.ID)
}

func TestInsertPersonWithCacheSuccess(t *testing.T) {
	cacherMock := &CacherMock{}
	querier := NewCachingQuerier(&QuerierMock{
		InsertPersonResult: Person{ID: 1, Name: "Test", Email: "email@email.com"},
	}, cacherMock)
	result, err := querier.InsertPerson(context.Background(), InsertPersonParams{
		Name:  "Test",
		Email: "email@email.com",
	})

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.ID)
	assert.Equal(t, "Test", result.Name)
	assert.Equal(t, "email@email.com", result.Email)
	assert.Equal(t, "person:1", cacherMock.SetStringKey)
	assert.Equal(t, `{"ID":1,"Name":"Test","Email":"email@email.com","CreatedAt":null,"UpdatedAt":null,"UpdateUser":""}`, cacherMock.SetStringValue)
}

func TestInsertPersonWithCacheFail(t *testing.T) {
	querier := NewCachingQuerier(&QuerierMock{
		InsertPersonResult: Person{ID: 1, Name: "Test", Email: "email@email.com"},
	}, &CacherMock{
		SetStringError: fmt.Errorf("error"),
	})
	result, err := querier.InsertPerson(context.Background(), InsertPersonParams{
		Name:  "Test",
		Email: "email@email.com",
	})

	assert.NotNil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1), result.ID)
	assert.Equal(t, "Test", result.Name)
	assert.Equal(t, "email@email.com", result.Email)
}

func TestUpdatePersonWithCacheSuccess(t *testing.T) {
	cacherMock := &CacherMock{}
	querier := NewCachingQuerier(&QuerierMock{
		UpdatePersonResult: 1,
	}, cacherMock)
	result, err := querier.UpdatePerson(context.Background(), UpdatePersonParams{
		ID:    1,
		Name:  "Test",
		Email: "email@email.com",
	})

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result)
	assert.Equal(t, "person:1", cacherMock.SetStringKey)
	assert.Equal(t, `{"ID":1,"Name":"Test","Email":"email@email.com","CreatedAt":"0001-01-01T00:00:00","UpdatedAt":"0001-01-01T00:00:00","UpdateUser":""}`, cacherMock.SetStringValue)
}

func TestUpdatePersonWithCacheFail(t *testing.T) {
	querier := NewCachingQuerier(&QuerierMock{
		UpdatePersonResult: 1,
	}, &CacherMock{
		SetStringError: fmt.Errorf("error"),
	})
	result, err := querier.UpdatePerson(context.Background(), UpdatePersonParams{
		ID:    1,
		Name:  "Test",
		Email: "email@email.com",
	})

	assert.NotNil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result)
}

func TestDeletePersonWithCacheSuccess(t *testing.T) {
	cacherMock := &CacherMock{}
	querier := NewCachingQuerier(&QuerierMock{
		DeletePersonResult: 1,
	}, cacherMock)
	result, err := querier.DeletePerson(context.Background(), 1)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result)
	assert.Equal(t, "person:1", cacherMock.DeleteKeyKey)
}

func TestDeletePersonWithCacheFail(t *testing.T) {
	querier := NewCachingQuerier(&QuerierMock{
		DeletePersonResult: 1,
	}, &CacherMock{
		DeleteKeyError: fmt.Errorf("error"),
	})
	result, err := querier.DeletePerson(context.Background(), 1)

	assert.NotNil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result)
}
