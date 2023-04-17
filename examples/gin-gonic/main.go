package main

import (
	"net/http"
	"reflect"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jakoblorz/scf/examples/gin-gonic/api"
	"github.com/jakoblorz/scf/spec"
	"github.com/mitchellh/mapstructure"
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

func main() {
	validate := validator.New()

	routes := api.GetRoutes()

	r := gin.Default()
	for _, endpointPtr := range routes {
		endpoint := *endpointPtr

		decodeParameters := func(c *gin.Context) bool { return true }
		if endpoint.Parameters != nil {
			decodeParameters = func(c *gin.Context) bool {
				params := map[string]string{}
				for _, param := range c.Params {
					params[param.Key] = param.Value
				}

				paramsStruct := safePtrClone(endpoint.Parameters)
				decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
					TagName: "json",
					Result:  paramsStruct,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, err.Error())
					return false
				}
				if err := decoder.Decode(params); err != nil {
					c.JSON(http.StatusInternalServerError, err.Error())
					return false
				}
				if err := validate.Struct(paramsStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return false
				}

				api.PutParams(c, paramsStruct)
				return true
			}
		}

		decodeQuery := func(c *gin.Context) bool { return true }
		if endpoint.Query != nil {
			decodeQuery = func(c *gin.Context) bool {
				queryStruct := safePtrClone(endpoint.Query)

				if err := c.ShouldBindQuery(queryStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return false
				}
				if err := validate.Struct(queryStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return false
				}

				api.PutQuery(c, queryStruct)
				return true
			}
		}

		decodePayload := func(c *gin.Context) bool { return true }
		if endpoint.Payload != nil && endpoint.Method != http.MethodGet {

			bodyForMediaType := make(map[string]interface{})
			for _, body := range endpoint.Payload {
				bodyForMediaType[body.MediaType] = body.Value
			}

			decodePayload = func(c *gin.Context) bool {
				bodyAnnotation, ok := bodyForMediaType[c.ContentType()]
				if !ok {
					c.JSON(http.StatusUnsupportedMediaType, "unsupported media type")
					return false
				}

				bodyStruct := safePtrClone(bodyAnnotation)
				if err := c.ShouldBindJSON(bodyStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return false
				}
				if err := validate.Struct(bodyStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return false
				}

				api.PutPayload(c, bodyStruct)
				return true
			}
		}

		r.Handle(endpoint.Method, URLParamsRegex.ReplaceAllString(endpoint.Path, ":$1"), func(c *gin.Context) {
			if !decodeParameters(c) {
				return
			}
			if !decodeQuery(c) {
				return
			}
			if !decodePayload(c) {
				return
			}

			endpoint.Handler(c)
		})
	}

	schema := spec.Generate(routes)
	r.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, schema)
	})

	r.Run()
}
