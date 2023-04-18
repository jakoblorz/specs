package query

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

func Decorate(c *fiber.Ctx, param interface{}) {
	ctx := c.UserContext()
	c.SetUserContext(context.WithValue(ctx, "query", param))
}

func Resolve[T interface{}](c *fiber.Ctx) *T {
	val := c.UserContext().Value("query")
	if val == nil {
		return new(T)
	}
	return val.(*T)
}
