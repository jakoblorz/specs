package main

import (
	"net/http"
	"reflect"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jakoblorz/scf"
	"github.com/jakoblorz/scf/examples/gin-gonic/api"
	"github.com/jakoblorz/scf/spec"
	"github.com/mitchellh/mapstructure"
)

var (
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

	r := gin.Default()
	for _, endpointPtr := range *api.Router() {
		endpoint := *endpointPtr
		r.Handle(endpoint.Method, URLParamsRegex.ReplaceAllString(endpoint.Path, ":$1"), func(c *gin.Context) {
			if endpoint.Parameters != nil {
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
					return
				}
				if err := decoder.Decode(params); err != nil {
					c.JSON(http.StatusInternalServerError, err.Error())
					return
				}
				if err := validate.Struct(paramsStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return
				}

				api.PutParams(c, paramsStruct)
			}

			if endpoint.Query != nil {
				queryStruct := safePtrClone(endpoint.Query)

				if err := c.ShouldBindQuery(queryStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return
				}
				if err := validate.Struct(queryStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return
				}

				api.PutQuery(c, queryStruct)
			}

			if endpoint.Payload != nil && endpoint.Method != http.MethodGet {
				mediaType := c.ContentType()

				var bodyAnnotation *scf.Body
				for _, annotation := range endpoint.Payload {
					if annotation.MediaType == mediaType {
						bodyAnnotation = &annotation
						break
					}
				}
				if bodyAnnotation == nil {
					c.JSON(http.StatusUnsupportedMediaType, "unsupported media type")
					return
				}

				bodyStruct := safePtrClone(bodyAnnotation.Value)
				if err := c.ShouldBindJSON(bodyStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return
				}
				if err := validate.Struct(bodyStruct); err != nil {
					c.JSON(http.StatusBadRequest, err.Error())
					return
				}

				api.PutPayload(c, bodyStruct)
			}

			endpoint.Handler(c)
		})
	}

	schema := spec.Generate(*api.Router())
	r.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, schema)
	})

	r.Run()
}
