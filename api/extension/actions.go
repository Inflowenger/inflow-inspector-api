package extensionControllers

import (
	"strings"
	"time"

	svcHandler "github.com/Inflowenger/inflow-fusion/svcHandler"
	"github.com/Inflowenger/inflow-inspector-api/etc"
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/Inflowenger/inflow-inspector-api/repository"
	"github.com/gofiber/fiber/v3"
)

func addNewExt(c fiber.Ctx) error {
	input := models.ExtensionRecord{}
	if err := c.Bind().Body(&input); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	if input.ID == "" {
		input.CreatedAt = time.Now().Unix()
	}
	input.UpdatedAt = time.Now().Unix()
	if strings.TrimSpace(input.Name) == "" {
		input.Name = "NoName"
	}
	err := repository.UpsertExtension(&input)
	if err != nil {
		return etc.Send(c, fiber.StatusInternalServerError, nil, models.ErrorResponse{Code: fiber.ErrInternalServerError.Code, Message: fiber.ErrInternalServerError.Message})

	}
	return etc.Send(c, fiber.StatusOK, input, nil)
}

func list(c fiber.Ctx) error {
	q := models.PaginationParams{}
	if err := c.Bind().Query(&q); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	l, cursor, err := repository.GetExtensionList(q.Cursor, int(q.PerPage))
	if err != nil {
		return etc.Send(c, fiber.StatusInternalServerError, nil, models.ErrorResponse{Code: fiber.ErrInternalServerError.Code, Message: fiber.ErrInternalServerError.Message})
	}
	return etc.Send(c, fiber.StatusOK, map[string]any{"list": l, "next": cursor}, err)
}

func deleteExtById(c fiber.Ctx) error {
	extId := c.Params("extId")
	if !strings.HasPrefix(extId, repository.EXTENSION_INDEX_PREFIX) {
		extId = repository.ExtensionIndexByString(extId)
	}
	err := repository.Delete(extId)
	if err != nil {
		return etc.Send(c, fiber.StatusInternalServerError, nil, models.ErrorResponse{Code: fiber.ErrInternalServerError.Code, Message: fiber.ErrInternalServerError.Message})
	}
	return etc.Send(c, fiber.StatusAccepted, models.Response{Data: map[string]any{"extensionId": extId}}, nil)

}
func getExtensionById(c fiber.Ctx) error {
	extId := c.Params("extId")
	if !strings.HasPrefix(extId, repository.EXTENSION_INDEX_PREFIX) {
		extId = repository.ExtensionIndexByString(extId)
	}
	ext, err := repository.GetExtensionById(extId)
	if err != nil {
		return etc.Send(c, fiber.StatusNotFound, nil, models.ErrorResponse{Code: fiber.ErrNotFound.Code, Message: "given extension id not found or error occured with" + err.Error()})

	}
	return etc.Send(c, fiber.StatusOK, ext, nil)
}
func listOfExtHandlers(c fiber.Ctx) error {
	list := svcHandler.GetAllSvcs()
	return etc.Send(c, fiber.StatusOK, list, nil)
}
