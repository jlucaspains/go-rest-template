package db

import (
	"context"
	"fmt"
	"goapi-template/cache"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
)

type CachingQuerier struct {
	Queries Querier
	Cache   cache.Cacher
}

func (c *CachingQuerier) GetPeople(ctx context.Context) ([]Person, error) {
	cached, err := cache.GetObject[[]Person](c.Cache, ctx, "person:all")

	if err != nil {
		slog.Error("Error getting person by id from cache", "error", err)
	}

	if cached != nil {
		return *cached, nil
	}

	result, err := c.Queries.GetPeople(ctx)

	cache.SetObject(c.Cache, ctx, "person:all", &result)

	return result, err
}

func (c *CachingQuerier) GetPersonById(ctx context.Context, id int32) (Person, error) {
	cached, err := cache.GetObject[Person](c.Cache, ctx, fmt.Sprintf("person:%d", id))

	if err != nil {
		slog.Error("Error getting person by id from cache", "error", err)
	}

	if cached != nil {
		return *cached, nil
	}

	result, err := c.Queries.GetPersonById(ctx, id)

	cache.SetObject(c.Cache, ctx, fmt.Sprintf("person:%d", id), &result)

	return result, err
}

func (c *CachingQuerier) InsertPerson(ctx context.Context, arg InsertPersonParams) (Person, error) {
	person, err := c.Queries.InsertPerson(ctx, arg)

	if err != nil {
		return person, err
	}

	err = cache.SetObject(c.Cache, ctx, fmt.Sprintf("person:%d", person.ID), &person)

	if err != nil {
		slog.Error("Error setting person by id into cache", "error", err)
	}

	return person, err
}

func (c *CachingQuerier) UpdatePerson(ctx context.Context, arg UpdatePersonParams) (int64, error) {
	personId, err := c.Queries.UpdatePerson(ctx, arg)

	if err != nil {
		return personId, err
	}

	cacheObj := &Person{
		ID:         arg.ID,
		Name:       arg.Name,
		Email:      arg.Email,
		UpdateUser: arg.UpdateUser,
		CreatedAt:  pgtype.Timestamp{Time: arg.CreatedAt.Time, Valid: true},
		UpdatedAt:  pgtype.Timestamp{Time: arg.UpdatedAt.Time, Valid: true},
	}

	err = cache.SetObject(c.Cache, ctx, fmt.Sprintf("person:%d", arg.ID), cacheObj)

	if err != nil {
		slog.Error("Error setting person by id into cache", "error", err)
	}

	return personId, err
}

func (c *CachingQuerier) DeletePerson(ctx context.Context, id int32) (int64, error) {
	personId, err := c.Queries.DeletePerson(ctx, id)

	if err != nil {
		return personId, err
	}

	err = c.Cache.DeleteKey(ctx, fmt.Sprintf("person:%d", id))

	if err != nil {
		slog.Error("Error deleting person by id from cache", "error", err)
	}

	return personId, err
}

func (c *CachingQuerier) PingDb(ctx context.Context) (int32, error) {
	return c.Queries.PingDb(ctx)
}

func NewCachingQuerier(querier Querier, cacher cache.Cacher) *CachingQuerier {
	return &CachingQuerier{Queries: querier, Cache: cacher}
}
