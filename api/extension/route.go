package extensionControllers

import (
	"github.com/Inflowenger/inflow-inspector-api/etc"
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/gofiber/fiber/v3"
	validation "github.com/mehdi-shokohi/fiberValidation"
)

func Register(api fiber.Router) {
	extGroup := api.Group("extension")
	extGroup.Use(etc.HS256SecKeyHandler())

	extGroup.Post("", validation.ValidateBodyAs[models.ExtensionRecord](), addNewExt)
	extGroup.Get("", list)
	extGroup.Get("/id/:extId", getExtensionById)
	extGroup.Delete("/id/:extId", deleteExtById)
	extGroup.Get("/extrinsics", listOfExtHandlers)

	//Plugins
	pluginGroup := extGroup.Group("plugin")
	pluginGroup.Post("/cred", validation.ValidateBodyAs[models.CredRequest](), getPluginCred)

}
