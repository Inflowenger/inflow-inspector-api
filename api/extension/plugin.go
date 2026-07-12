package extensionControllers

import (
	"fmt"

	InfraSpaces "github.com/Inflowenger/inflow-fusion/spaces"
	"github.com/Inflowenger/inflow-inspector-api/etc"
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/gofiber/fiber/v3"
	"github.com/nats-io/nkeys"
)

func getPluginCred(c fiber.Ctx) error {
	input := models.CredRequest{}
	if err := c.Bind().Body(&input); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	spaceSeed := ""
	if input.SpaceId == "" {
		acc, err := InfraSpaces.GetPluginBuiltinAccount()
		if err != nil {
			return etc.Send(c, fiber.StatusInternalServerError, nil, err)
		}
		spaceSeed = acc.Seed
	} else {
		acc, err := InfraSpaces.GetAccountById(input.SpaceId)
		if err != nil {
			return etc.Send(c, fiber.StatusInternalServerError, nil, err)
		}
		spaceSeed = acc.Seed
	}
	kp, err := nkeys.FromSeed([]byte(spaceSeed))
	if err != nil {
		return etc.Send(c, fiber.StatusNotAcceptable, nil, fmt.Errorf("invalid keys"))
	}
	pub, _ := kp.PublicKey()
	var cred string
	switch models.AccessCredType(input.Access) {
	case models.MultiPluginAccess:
		ucred, err := InfraSpaces.CreateUserCredential(spaceSeed, InfraSpaces.PluginCredentialOpenPermission(input.Name, input.PluginId, pub))
		if err != nil {
			return etc.Send(c, fiber.StatusNotAcceptable, nil, fmt.Errorf("error occurred in create access token"))

		}
		cred = ucred.Base64Cred
	case models.StrictAccess:

		ucred, err := InfraSpaces.CreateUserCredential(spaceSeed, InfraSpaces.PluginCredentialStrictPermission(input.Name, input.PluginId, pub))
		if err != nil {
			return etc.Send(c, fiber.StatusNotAcceptable, nil, fmt.Errorf("error occurred in create access token"))

		}
		cred = ucred.Base64Cred
	}
	envTemplate := `
		INFRA_CRED={cred}
		INFRA_URL={url}
		PLUGIN_ID={pluginId}
	`
	env := etc.ReplaceByMapString(envTemplate, map[string]any{"cred": cred, "pluginId": input.PluginId, "url": "infra:4222"})
	return etc.Send(c, fiber.StatusOK, map[string]any{"env": env, "cred": cred}, nil)
}
