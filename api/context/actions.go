package contextControllers

import (
	"strings"
	"time"

	"github.com/Inflowenger/inflow-inspector-api/etc"
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/Inflowenger/inflow-inspector-api/repository"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
)

func addNewContext(c fiber.Ctx) error {
	input := models.ContextRecord{}
	if err := c.Bind().Body(&input); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	if input.ID == "" {
		input.CreatedAt = time.Now().Unix()
	}
	input.UpdatedAt = time.Now().Unix()
	input.UpdatedBy = models.LastChange{By: models.ByAPI}
	if strings.TrimSpace(input.Title) == "" {
		input.Title = "untitled context"
	}
	do := map[string]any{}
	err := sonic.Unmarshal([]byte(input.Context), &do)
	if err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: "invaid context data - context is a json object"})
	}

	err = repository.UpsertContext(&input)
	if err != nil {
		return etc.Send(c, fiber.StatusInternalServerError, nil, models.ErrorResponse{Code: fiber.ErrInternalServerError.Code, Message: fiber.ErrInternalServerError.Message})

	}
	return etc.Send(c, fiber.StatusOK, input, nil)
}

func getContextById(c fiber.Ctx) error {
	contextId := c.Params("contextId")
	if !strings.HasPrefix(contextId, repository.CONTEXT_INDEX_PREFIX) {
		contextId = repository.ContextIndexByString(contextId)
	}
	context := repository.GetContextById(contextId)
	if context == nil {
		return etc.Send(c, fiber.StatusNotFound, nil, models.ErrorResponse{Code: fiber.ErrNotFound.Code, Message: "given context id not found"})

	}
	return etc.Send(c, fiber.StatusOK, context, nil)
}
func deleteContextById(c fiber.Ctx) error {
	contextId := c.Params("contextId")
	if !strings.HasPrefix(contextId, repository.CONTEXT_INDEX_PREFIX) {
		contextId = repository.ContextIndexByString(contextId)
	}
	err := repository.Delete(contextId)
	if err != nil {
		return etc.Send(c, fiber.StatusInternalServerError, nil, models.ErrorResponse{Code: fiber.ErrInternalServerError.Code, Message: fiber.ErrInternalServerError.Message})
	}
	return etc.Send(c, fiber.StatusAccepted, map[string]any{"contextId": contextId}, nil)

}
func list(c fiber.Ctx) error {
	q := models.PaginationParams{}
	if err := c.Bind().Query(&q); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	l, cursor, err := repository.GetContextList(q.Cursor, int(q.PerPage))
	if err != nil {
		return etc.Send(c, fiber.StatusInternalServerError, nil, models.ErrorResponse{Code: fiber.ErrInternalServerError.Code, Message: fiber.ErrInternalServerError.Message})
	}
	return etc.Send(c, fiber.StatusOK, map[string]any{"list": l, "next": cursor}, err)
}
