package api

import (
	"context"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jakoblorz/specs"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"reflect"
	"regexp"
)

var (
	// URLParamsRegex is a regular expression to match URL parameters in the openapi spec format (e.g. {id}).
	// It allows to replace the parameters with the gin-gonic format (e.g. :id).
	URLParamsRegex = regexp.MustCompile(`\{([a-zA-Z0-9]+)\}`)
)

// safePtrClone returns a pointer to a new instance of the given type.
func safePtrClone(v interface{}) interface{} {
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		return reflect.New(reflect.TypeOf(v).Elem()).Interface()
	}
	return reflect.New(reflect.TypeOf(v)).Interface()
}

func decorateQuery(c *fiber.Ctx, param interface{}) {
	ctx := c.UserContext()
	c.SetUserContext(context.WithValue(ctx, "query", param))
}

func resolveQuery[T interface{}](c *fiber.Ctx) *T {
	val := c.UserContext().Value("query")
	if val == nil {
		return new(T)
	}
	return val.(*T)
}

func decoratePayload(c *fiber.Ctx, param interface{}) {
	ctx := c.UserContext()
	c.SetUserContext(context.WithValue(ctx, "payload", param))
}

func resolvePayload[T interface{}](c *fiber.Ctx) *T {
	val := c.UserContext().Value("payload")
	if val == nil {
		return new(T)
	}
	return val.(*T)
}

func decorateParams(c *fiber.Ctx, param interface{}) {
	ctx := c.UserContext()
	c.SetUserContext(context.WithValue(ctx, "params", param))
}

func resolveParams[T interface{}](c *fiber.Ctx) *T {
	val := c.UserContext().Value("params")
	if val == nil {
		return new(T)
	}
	return val.(*T)
}

var (
	router = specs.NewRegistry[fiber.Handler]()
)

func Mount(app *fiber.App) {
	validate := validator.New()

	for _, endpointPtr := range router.Eject() {
		endpoint := *endpointPtr

		decodeParameters := func(c *fiber.Ctx) bool { return true }
		if endpoint.Parameters != nil {
			decodeParameters = func(c *fiber.Ctx) bool {
				paramsStruct := safePtrClone(endpoint.Parameters)
				decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
					TagName: "json",
					Result:  paramsStruct,
				})
				if err != nil {
					c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"message": err.Error(),
					})
					return false
				}
				if err := decoder.Decode(c.AllParams()); err != nil {
					c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"message": err.Error(),
					})
					return false
				}
				if err := validate.Struct(paramsStruct); err != nil {
					c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"message": err.Error(),
					})
					return false
				}

				decorateParams(c, paramsStruct)
				return true
			}
		}

		decodeQuery := func(c *fiber.Ctx) bool { return true }
		if endpoint.Query != nil {
			decodeQuery = func(c *fiber.Ctx) bool {
				queryStruct := safePtrClone(endpoint.Query)

				if err := c.QueryParser(queryStruct); err != nil {
					c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"message": err.Error(),
					})
					return false
				}

				if err := validate.Struct(queryStruct); err != nil {
					c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"message": err.Error(),
					})
					return false
				}

				decorateQuery(c, queryStruct)
				return true
			}
		}

		decodePayload := func(c *fiber.Ctx) bool { return true }
		if endpoint.Payload != nil && endpoint.Method != http.MethodGet {
			decodePayload = func(c *fiber.Ctx) bool {
				payloadStruct := safePtrClone(endpoint.Payload)

				if err := c.BodyParser(payloadStruct); err != nil {
					c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"message": err.Error(),
					})
					return false
				}

				if err := validate.Struct(payloadStruct); err != nil {
					c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"message": err.Error(),
					})
					return false
				}

				decoratePayload(c, payloadStruct)
				return true
			}
		}

		app.Add(endpoint.Method, URLParamsRegex.ReplaceAllString(endpoint.Path, ":$1"), func(c *fiber.Ctx) error {
			if !decodeParameters(c) {
				return nil
			}
			if !decodeQuery(c) {
				return nil
			}
			if !decodePayload(c) {
				return nil
			}

			if err := endpoint.Handler(c); err != nil {
				c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"message": err.Error(),
				})
			}
			return nil
		})
	}
}

func Annotate(t *openapi3.T) {
	router.Annotate(t)
}
