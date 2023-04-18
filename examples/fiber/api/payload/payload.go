package payload

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

func Decorate(c *fiber.Ctx, param interface{}) {
	ctx := c.UserContext()
	c.SetUserContext(context.WithValue(ctx, "payload", param))
}

func Resolve[T interface{}](c *fiber.Ctx) *T {
	val := c.UserContext().Value("payload")
	if val == nil {
		return new(T)
	}
	return val.(*T)
}
