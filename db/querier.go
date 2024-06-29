package db

import "context"

type Querier interface {
	GetPeople(ctx context.Context) ([]Person, error)
	GetPersonById(ctx context.Context, id int32) (Person, error)
	InsertPerson(ctx context.Context, arg InsertPersonParams) (Person, error)
	UpdatePerson(ctx context.Context, arg UpdatePersonParams) (Person, error)
	DeletePerson(ctx context.Context, id int32) (int64, error)
	PingDb(ctx context.Context) (int32, error)
}
