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

func safePtrClone(v interface{}) interface{} {
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		return reflect.New(reflect.TypeOf(v).Elem()).Interface()
	}
	return reflect.New(reflect.TypeOf(v)).Interface()
}

func main() {
	validate := validator.New()

	r := gin.Default()
	for _, endpoint := range *api.Router() {
		r.Handle(endpoint.Method, URLParamsRegex.ReplaceAllString(endpoint.Path, ":$1"), func(c *gin.Context) {
			if endpoint.Parameters != nil {
				params := map[string]string{}
				for _, param := range c.Params {
					params[param.Key] = param.Value
				}

				paramsStruct := safePtrClone(endpoint.Parameters)
				mapstructure.Decode(params, paramsStruct)

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

				api.PutQuery(c, queryStruct)
			}

			if endpoint.Payload != nil {
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
