package contextControllers

import (
	"github.com/Inflowenger/inflow-inspector-api/etc"
	"github.com/gofiber/fiber/v3"
)

func Register(api fiber.Router) {
	contextGroup := api.Group("context")
	contextGroup.Use(etc.HS256SecKeyHandler())

	contextGroup.Post("", addNewContext)
	contextGroup.Get("", list)
	contextGroup.Get("/id/:contextId", getContextById)
	contextGroup.Delete("/id/:contextId", deleteContextById)

}
