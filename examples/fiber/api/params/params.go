package params

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

func Decorate(c *fiber.Ctx, param interface{}) {
	ctx := c.UserContext()
	c.SetUserContext(context.WithValue(ctx, "params", param))
}

func Resolve[T interface{}](c *fiber.Ctx) *T {
	val := c.UserContext().Value("params")
	if val == nil {
		return new(T)
	}
	return val.(*T)
}
