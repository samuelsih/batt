package batt

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestHandler_EmptyRequest(t *testing.T) {
	type TestRequest struct {
	}

	type TestResponse struct {
		Message string `json:"message"`
	}

	businessLogic := func(ctx context.Context, r TestRequest) (TestResponse, error) {
		res := TestResponse{
			Message: "Hello world",
		}

		return res, nil
	}

	app := fiber.New(fiber.Config{Immutable: true})
	app.Get("/hello", Handler(businessLogic, 200))

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode)
	require.JSONEq(t, `{"message": "Hello world"}`, string(body))
}

func TestHandler_WithRequest(t *testing.T) {
	type TestStruct struct {
		Page int    `query:"page"`
		ID   string `params:"id"`
		Name string `json:"name"`
	}

	type TestResponse struct {
		Page      int    `json:"page"`
		ID        string `json:"id"`
		Name      string `json:"name"`
		Something string `json:"something"`
	}

	businessLogic := func(ctx context.Context, r TestStruct) (TestResponse, error) {
		return TestResponse{
			Page:      r.Page,
			ID:        r.ID,
			Name:      r.Name,
			Something: "something",
		}, nil
	}

	buf, err := json.Marshal(struct {
		Name string `json:"name"`
	}{Name: "Hello"})

	require.NoError(t, err)

	app := fiber.New(fiber.Config{Immutable: true})
	app.Post("/user/:id", Handler(businessLogic, 200))

	req := httptest.NewRequest(http.MethodPost, "/user/123?page=1", bytes.NewBuffer(buf))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode)
	require.JSONEq(t, `{"page": 1, "id": "123", "name": "Hello", "something": "something"}`, string(body))
}

func Test_CodeBeforeClosure(t *testing.T) {
	counter := 0
	msg := "not triggered"

	f := func() func() {
		counter++
		return func() {
			msg = strconv.Itoa(counter)
		}
	}

	trigger := f()
	for range 3 {
		trigger()
	}

	require.Equal(t, 1, counter)
	require.Equal(t, "1", msg)
}
