package main

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/jakoblorz/specs/examples/fiber/api"
	"net/http"
)

func main() {
	app := fiber.New()
	api.Mount(app)

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

	app.Get("/openapi.json", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(t)
	})

	app.Listen(":8080")
}
