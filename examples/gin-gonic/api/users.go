package api

import (
	"github.com/gin-gonic/gin"
)

func init() {
	router.POST("/api/users", Handle_CreateNewUserRequest).
		Title("Creates a new user").
		Description("Creates a new user").
		Tags("api", "users").
		Payload(CreateNewUserRequest{}).
		Response(201, User{}, "User created")
}

type CreateNewUserRequest struct {
	Name string `json:"name" validate:"required"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Nick string `json:"nick"`
}

func Handle_CreateNewUserRequest(c *gin.Context) {
	var (
		payload = GetPayload[CreateNewUserRequest](c)
	)

	c.JSON(201, User{
		ID:   "abc",
		Name: payload.Name,
		Nick: payload.Name + "nick",
	})
}

type DetailedURLParameters struct {
	UserID string `json:"id"`
}

func init() {
	router.PUT("/api/users/{id}", Handle_UpdateUserRequest).
		Title("Update a User").
		Description("Update a user").
		Tags("api", "users").
		Parameters(DetailedURLParameters{}).
		Payload(UpdateUserRequest{}).
		Response(200, UpdateUserResponse{}, "User updated")
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UpdateUserResponse struct {
	Name string `json:"name"`
	Nick string `json:"nick"`
}

func Handle_UpdateUserRequest(c *gin.Context) {
	var (
		params  = GetParams[DetailedURLParameters](c)
		payload = GetPayload[UpdateUserRequest](c)
	)

	c.JSON(200, User{
		ID:   params.UserID,
		Name: payload.Name,
		Nick: payload.Name + "nick",
	})
}
