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

func Handle_CreateNewUserRequest(c *gin.Context) {
	var (
		payload = GetPayload[CreateNewUserRequest](c)
	)

	c.JSON(201, CreateNewUserResponse{
		ID:   "abc",
		Name: payload.Name,
		Nick: payload.Name + "nick",
	})
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
	Docs  []GetUserResponse `json:"items"`
}

func Handle_GetUsersRequest(c *gin.Context) {
	var (
		query = GetQuery[GetUsersQueryParameters](c)
	)

	c.JSON(200, GetUsersResponse{
		Total: query.Limit,
		Docs:  []GetUserResponse{},
	})
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
		Response(200, GetUserResponse{}, "User found")
}

type GetUserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Nick string `json:"nick"`
}

func Handle_GetUserRequest(c *gin.Context) {
	var (
		params = GetParams[DetailedURLParameters](c)
	)

	c.JSON(200, GetUserResponse{
		ID:   params.UserID,
		Name: "John Doe",
		Nick: "John Doe" + "nick",
	})
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
	ID   string `json:"id"`
	Name string `json:"name"`
	Nick string `json:"nick"`
}

func Handle_UpdateUserRequest(c *gin.Context) {
	var (
		params  = GetParams[DetailedURLParameters](c)
		payload = GetPayload[UpdateUserRequest](c)
	)

	c.JSON(200, UpdateUserResponse{
		ID:   params.UserID,
		Name: payload.Name,
		Nick: payload.Name + "nick",
	})
}
