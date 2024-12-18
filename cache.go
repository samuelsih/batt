package batt

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/rueidis"
)

type InMem interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

var (
	inmem        InMem
	StoreForever = time.Duration(0)

	// interface check build time
	_ InMem = &redis{}
)

var (
	ErrInMemNotFound = errors.New("item not found")
)

func InMemGet(ctx context.Context, key string) (string, error) {
	return inmem.Get(ctx, key)
}

func InMemSet(ctx context.Context, key string, value string, ttl time.Duration) error {
	return inmem.Set(ctx, key, value, ttl)
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

func InMemSetJSON[T any](ctx context.Context, key string, value T, ttl time.Duration) error {
	t, err := json.Marshal(value)
	if err != nil {
		println("ish error", err.Error())
		return err
	}

	err = inmem.Set(ctx, key, string(t), ttl)
	return err
}

type redis struct {
	client rueidis.Client
}

func (r *redis) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Do(ctx, r.client.B().Get().Key(key).Build()).ToString()
	if err == nil {
		return result, nil
	}

	if r.emptyGetError(err) {
		return "", ErrInMemNotFound
	}

	return result, err
}

func (r *redis) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl == StoreForever {
		err := r.client.Do(ctx, r.client.B().Set().Key(key).Value(value).Nx().Build()).Error()
		return err
	}

	err := r.client.Do(ctx, r.client.B().Set().Key(key).Value(value).Ex(ttl).Build()).Error()
	return err
}

func (redis) emptyGetError(err error) bool {
	return err != nil && (err.Error() == `redis nil message` || err.Error() == `redis: nil`)
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
