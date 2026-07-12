package flowControllers

import (
	"github.com/Inflowenger/inflow-inspector-api/etc"
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/gofiber/contrib/v3/socketio"
	validation "github.com/mehdi-shokohi/fiberValidation"

	"github.com/gofiber/fiber/v3"
)

func Register(api fiber.Router) {
	flowGroup := api.Group("flow")
	flowGroup.Use(etc.HS256SecKeyHandler())
	flowGroup.Post("", addNewFlow)
	flowGroup.Get("", list)
	flowGroup.Get("/id/:flowId", getFlowById)
	flowGroup.Delete("/id/:flowId", deleteFlowById)

	// process requests
	procGroup := api.Group("ps")
	procGroup.Use(etc.HS256SecKeyHandler())

	procGroup.Post("", validation.ValidateBodyAs[models.ProcessRequestInput](), newProcess)
	procGroup.Post("/compile", validation.ValidateBodyAs[models.ProcessRequestInput](), compile)

	procGroup.Post("/stop/:pid", stopByPid)

	// websockets logs
	// LoadEventHandlers() is registered lazily in GetWsSessions().
	api.Get("/ws/:id", wsAuthHandler, socketio.New(wshandler))
	api.Post("/sendto", sendmessageto)
}
