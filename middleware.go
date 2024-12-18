package batt

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type authKey struct{}

type ErrAuthBearerHandler struct {
	OnHeaderMissing func(*fiber.Ctx) error
	OnInvalidUser   func(*fiber.Ctx) error
}

var (
	CtxAuthKey                  = authKey{}
	DefaultErrAuthBearerHandler = ErrAuthBearerHandler{
		OnHeaderMissing: func(c *fiber.Ctx) error {
			return c.Status(401).JSON(fiber.Map{"message": "Unauthenticated"})
		},
		OnInvalidUser: func(c *fiber.Ctx) error {
			return c.Status(401).JSON(fiber.Map{"message": "Unauthenticated"})
		},
	}
)

func AuthBearer[T any](handle func(ctx context.Context, token string) (T, bool)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization, "")
		if len(header) == 0 {
			return DefaultErrAuthBearerHandler.OnHeaderMissing(c)
		}

		components := strings.SplitN(header, " ", 2)
		headerName, token := components[0], components[1]
		if headerName != "Bearer" {
			return DefaultErrAuthBearerHandler.OnHeaderMissing(c)
		}

		user, ok := handle(c.UserContext(), token)
		if !ok {
			return DefaultErrAuthBearerHandler.OnInvalidUser(c)
		}

		c.SetUserContext(context.WithValue(context.Background(), CtxAuthKey, user))
		return c.Next()
	}
}

func GetAuthUser[T any](ctx context.Context) (T, bool) {
	user, ok := ctx.Value(CtxAuthKey).(T)
	return user, ok
}
