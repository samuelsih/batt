package batt

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestAuthBearer(t *testing.T) {
	type user struct {
		ID string
	}

	auth := AuthBearer(func(ctx context.Context, token string) (user, bool) {
		return user{ID: "1"}, true
	})

	app := fiber.New()
	app.Get("/foo", auth, func(c *fiber.Ctx) error {
		u, ok := GetAuthUser[user](c.UserContext())
		require.True(t, ok)
		return c.SendString(u.ID)
	})

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Add("Authorization", "Bearer 123123")
	
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, "1", string(body))
}
