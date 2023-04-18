package api

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jakoblorz/scf"
	"github.com/jakoblorz/scf/examples/fiber/api/params"
	"github.com/jakoblorz/scf/examples/fiber/api/payload"
	"github.com/jakoblorz/scf/examples/fiber/api/query"
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

var (
	router = scf.NewRegistry[fiber.Handler]()
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

				params.Decorate(c, paramsStruct)
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

				query.Decorate(c, queryStruct)
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

				payload.Decorate(c, payloadStruct)
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
