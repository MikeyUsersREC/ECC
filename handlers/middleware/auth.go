package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// mikey u can use this later for validating shite

func ValidateRequest() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}
