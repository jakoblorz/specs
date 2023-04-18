package main

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jakoblorz/specs/examples/gin-gonic/api"
)

func main() {

	r := gin.Default()
	r.Use(cors.Default())
	api.Mount(r)

	t := new(openapi3.T)
	t.OpenAPI = "3.0.0"
	t.Info = &openapi3.Info{
		Title:   "Example API",
		Version: "1.0.0",
	}
	t.AddServer(&openapi3.Server{
		URL: "http://localhost:8080",
	})
	api.Annotate(t)

	r.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, t)
	})

	r.Run()
}
