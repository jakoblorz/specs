package api

import (
	"github.com/gofiber/fiber/v2"
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
	Age  int    `json:"age" validate:"required,gte=18"`
}

type CreateNewUserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func Handle_CreateNewUserRequest(c *fiber.Ctx) error {
	payload := resolvePayload[CreateNewUserRequest](c)

	c.Status(http.StatusCreated).JSON(CreateNewUserResponse{
		ID:   "abc",
		Name: payload.Name,
		Age:  payload.Age,
	})
	return nil
}

func init() {
	router.GET("/api/users", Handle_GetUsersRequest).
		Title("Get all users").
		Description("Get all users").
		Tags("api", "users").
		Query(GetUsersQueryParameters{}).
		Response(200, GetUsersResponse{}, "Users found")
}

type GetUsersQueryParameters struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type GetUsersResponse struct {
	Total int               `json:"total"`
	Docs  []GetUserResponse `json:"docs"`
}

func Handle_GetUsersRequest(c *fiber.Ctx) error {
	query := resolveQuery[GetUsersQueryParameters](c)

	c.Status(http.StatusOK).JSON(GetUsersResponse{
		Total: query.Limit,
		Docs:  []GetUserResponse{},
	})
	return nil
}

type DetailedURLParameters struct {
	UserID string `json:"id" validate:"required"`
}

func init() {
	router.GET("/api/users/{id}", Handle_GetUserRequest).
		Title("Get a User").
		Description("Get a user").
		Tags("api", "users").
		Parameters(DetailedURLParameters{}).
		Response(http.StatusOK, GetUserResponse{}, "User found")
}

type GetUserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func Handle_GetUserRequest(c *fiber.Ctx) error {
	params := resolveParams[DetailedURLParameters](c)

	c.Status(http.StatusOK).JSON(GetUserResponse{
		ID:   params.UserID,
		Name: "John Doe",
		Age:  18,
	})

	return nil
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
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"required,gte=18"`
}

type UpdateUserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func Handle_UpdateUserRequest(c *fiber.Ctx) error {
	params := resolveParams[DetailedURLParameters](c)
	payload := resolvePayload[UpdateUserRequest](c)

	c.Status(http.StatusOK).JSON(UpdateUserResponse{
		ID:   params.UserID,
		Name: payload.Name,
		Age:  payload.Age,
	})

	return nil
}
