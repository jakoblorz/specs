package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jakoblorz/scf"
	"github.com/jakoblorz/scf/spec"
)

var (
	reg = scf.Registry[gin.HandlerFunc]{}
)

func init() {
	reg.Bind(http.MethodPost, "/api/users", Handle_CreateNewUserRequest).
		Title("Creates a new user").
		Description("Creates a new user").
		Tags("api", "users").
		Consumes(CreateNewUserRequest{}, "application/json").
		Produces(201, User{}, "application/json").
		Build()
}

type CreateNewUserRequest struct {
	Name string `json:"name"`
}

type User struct {
	Name string `json:"name"`
	Nick string `json:"nick"`
}

func Handle_CreateNewUserRequest(c *gin.Context) {
	c.JSON(201, User{
		Name: "name",
		Nick: "nick",
	})
}

func main() {
	r := gin.Default()
	for _, endpoint := range reg {
		r.Handle(endpoint.Method, endpoint.Path, endpoint.Handler)
	}

	r.GET("/openapi.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, spec.Generate(reg))
	})

	r.Run()
}
