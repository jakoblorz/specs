package api

import (
	"github.com/gofiber/fiber/v2"
	payload "github.com/jakoblorz/scf/examples/fiber/api/payload"
	"net/http"
)

func init() {
	router.POST("/api/users", Handle_CreateNewUserRequest).
		Title("Creates a new user").
		Description("Creates a new user").
		Tags("api", "users").
		Payload(CreateNewUserRequest{}).
		Response(201, CreateNewUserResponse{}, "User created")
}

type CreateNewUserRequest struct {
	Name string `json:"name" validate:"required"`
}

type CreateNewUserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Nick string `json:"nick"`
}

func Handle_CreateNewUserRequest(c *fiber.Ctx) error {
	body := payload.Resolve[CreateNewUserRequest](c)

	c.Status(http.StatusCreated).JSON(CreateNewUserResponse{
		ID:   "abc",
		Name: body.Name,
		Nick: body.Name + "nick",
	})
	return nil
}
