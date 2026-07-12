package etc

import (
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/gofiber/fiber/v3"
)

func Send(c fiber.Ctx, code int, data any, error any) error {
	return c.Status(code).JSON(models.Response{Data: data, Error: error})

}
