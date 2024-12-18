package batt

import (
	"context"
	"reflect"
	"slices"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type NoParam = struct{}

type Response struct {
	Message string `json:"message"`
}

type BusinessLogic[Request, Response any] func(context.Context, Request) (Response, error)

var (
	validate      = validator.New()
	skipValidator = false
)

// SkipValidator determines if the request must passed go-playground/validator
func SkipValidator(want bool) {
	skipValidator = want
}

// BaseResponse is a function that returns api.Response without having to
// instantiate it over and over again.
func BaseResponse(msg string) Response {
	return Response{Message: msg}
}

// NoopResponse is a response that returns empty api.Response.
// The convention is api.NoopResponse() mus be called when err != nil in the response.
func NoopResponse() Response {
	return Response{}
}

func (r *Response) OK() {
	r.Message = "OK"
}

// HasRequestInput check if struct type has some property, not just plain struct{}{}.
func HasRequestInput[Request any]() bool {
	hasInput := false
	if reflect.TypeFor[Request]() != reflect.TypeOf(NoParam{}) {
		hasInput = true
	}

	return hasInput
}

// HasStructTags check if struct type contains at least one tag on its property.
func HasStructTags[Request any](tags ...string) bool {
	for field := range slices.Values(reflect.VisibleFields(reflect.TypeFor[Request]())) {
		for tag := range slices.Values(tags) {
			structTag := field.Tag.Get(tag)
			if structTag != "" {
				return true
			}
		}
	}

	return false
}

// Handler is wrapper for fiber.Handler that skips the presentation/serialization layer.
func Handler[Request, Response any](businessLogic BusinessLogic[Request, Response], successStatusCode int) fiber.Handler {
	hasInput := HasRequestInput[Request]()
	hasBody := HasStructTags[Request]("json", "form", "xml")

	return func(c *fiber.Ctx) error {
		var (
			req Request
			err error
		)

		hasQuery := len(c.Queries()) > 0
		hasParam := len(c.AllParams()) > 0

		if hasInput {
			if hasQuery {
				err = c.QueryParser(&req)
				if err != nil {
					return err
				}
			}

			if hasParam {
				err = c.ParamsParser(&req)
				if err != nil {
					return err
				}
			}

			if hasBody {
				if err := c.BodyParser(&req); err != nil {
					return err
				}

				if !skipValidator {
					err = validate.Struct(&req)
					if err != nil {
						return err
					}
				}
			}
		}

		res, err := businessLogic(c.UserContext(), req)
		if err != nil {
			return err
		}

		return c.Status(successStatusCode).JSON(res)
	}
}

func AuthUser[T any](ctx context.Context) T {
	user, _ := ctx.Value("user").(T)
	return user
}
