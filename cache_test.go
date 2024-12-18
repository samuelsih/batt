package batt

import (
	"context"
	"sync"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/redis/rueidis"
	"github.com/stretchr/testify/require"
)

func TestRedisImpl(t *testing.T) {
	r, cleanup, err := setupRedisClient()
	require.NoError(t, err)

	t.Cleanup(func() { cleanup() })

	c := redis{client: r}
	ctx := context.Background()

	t.Run("plain", func(t *testing.T) {
		err = c.Set(ctx, "foo", "bar", StoreForever)
		require.NoError(t, err)

		val, err := c.Get(ctx, "foo")
		require.NoError(t, err)
		require.Equal(t, "bar", val)
	})

	t.Run("json", func(t *testing.T) {
		type some struct{ Value string }

		var mu sync.Mutex
		t.Cleanup(func() { mu.Unlock() })

		mu.Lock()

		inmem = &c
		payload := some{Value: "something"}

		err = InMemSetJSON(ctx, "eyoo", payload, StoreForever)
		require.NoError(t, err)

		val, err := InMemGetJSON[some](ctx, "eyoo")
		require.NoError(t, err)
		require.EqualValues(t, payload.Value, val.Value)
	})
}

func setupRedisClient() (rueidis.Client, func(), error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, err
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, nil, err
	}

	resource, err := pool.Run("redis", "7-alpine", nil)
	if err != nil {
		return nil, nil, err
	}

	var c rueidis.Client

	if err = pool.Retry(func() error {
		addr := resource.GetHostPort("6379/tcp")
		c, err = rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{addr}})
		return err
	}); err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		pool.Purge(resource)
	}

	return c, cleanup, nil
}
