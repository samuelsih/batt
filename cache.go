package batt

import (
	"context"
	"encoding/json"

	"github.com/redis/rueidis"
)

type InMem interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}

var (
	inmem InMem
	_     InMem = &redis{}
)

func InMemGet(ctx context.Context, key string) (string, error) {
	return inmem.Get(ctx, key)
}

func InMemSet(ctx context.Context, key string, value string) error {
	return inmem.Set(ctx, key, value)
}

func InMemGetJSON[T any](ctx context.Context, key string) (T, error) {
	var t T

	str, err := inmem.Get(ctx, key)
	if err != nil {
		return t, err
	}

	err = json.Unmarshal([]byte(str), &t)
	if err != nil {
		return t, err
	}

	return t, nil
}

func InMemSetJSON[T any](ctx context.Context, key string, value T) error {
	t, err := json.Marshal(value)
	if err != nil {
		println("ish error", err.Error())
		return err
	}

	err = inmem.Set(ctx, key, string(t))
	return err
}

type redis struct {
	client rueidis.Client
}

func SetInMemRedis(username, password string, addr ...string) error {
	r, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: addr,
		Password:    password,
		Username:    username,
	})

	if err != nil {
		return err
	}

	client := redis{client: r}
	inmem = &client

	return nil
}

func (r *redis) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Do(ctx, r.client.B().Get().Key(key).Build()).ToString()
	return result, err
}

func (r *redis) Set(ctx context.Context, key string, value string) error {
	err := r.client.Do(ctx, r.client.B().Set().Key(key).Value(value).Build()).Error()
	return err
}
