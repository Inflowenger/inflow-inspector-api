package flowControllers

import (
	"strings"
	"time"

	"github.com/Inflowenger/inflow-inspector-api/etc"
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/Inflowenger/inflow-inspector-api/repository"
	"github.com/gofiber/fiber/v3"
)

func addNewFlow(c fiber.Ctx) error {
	input := models.FlowRecord{}
	if err := c.Bind().Body(&input); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	if input.ID == "" {
		input.CreatedAt = time.Now().Unix()
	}
	input.UpdatedAt = time.Now().Unix()
	if strings.TrimSpace(input.Title) == "" {
		input.Title = "untitled workflow"
	}
	err := repository.UpsertFlow(&input)
	if err != nil {
		return etc.Send(c, fiber.StatusInternalServerError, nil, models.ErrorResponse{Code: fiber.ErrInternalServerError.Code, Message: fiber.ErrInternalServerError.Message})

	}
	return etc.Send(c, fiber.StatusOK, input, nil)
}

func getFlowById(c fiber.Ctx) error {
	flowId := c.Params("flowId")
	if !strings.HasPrefix(flowId, repository.FLOW_INDEX_PREFIX) {
		flowId = repository.FlowIndexByString(flowId)
	}
	flow, err := repository.GetFlowById(flowId)
	if err != nil {
		return etc.Send(c, fiber.StatusNotFound, nil, models.ErrorResponse{Code: fiber.ErrNotFound.Code, Message: "given flow id not found or error occured with" + err.Error()})

	}
	return etc.Send(c, fiber.StatusOK, flow, nil)
}
func deleteFlowById(c fiber.Ctx) error {
	flowId := c.Params("flowId")
	if !strings.HasPrefix(flowId, repository.FLOW_INDEX_PREFIX) {
		flowId = repository.FlowIndexByString(flowId)
	}
	err := repository.Delete(flowId)
	if err != nil {
		return etc.Send(c, fiber.StatusInternalServerError, nil, models.ErrorResponse{Code: fiber.ErrInternalServerError.Code, Message: fiber.ErrInternalServerError.Message})
	}
	return etc.Send(c, fiber.StatusAccepted, models.Response{Data: map[string]any{"flowId": flowId}}, nil)

}
func list(c fiber.Ctx) error {
	q := models.PaginationParams{}
	if err := c.Bind().Query(&q); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	l, cursor, err := repository.GetFlowList(q.Cursor, int(q.PerPage))
	if err != nil {
		return etc.Send(c, fiber.StatusInternalServerError, nil, models.ErrorResponse{Code: fiber.ErrInternalServerError.Code, Message: fiber.ErrInternalServerError.Message})
	}
	return etc.Send(c, fiber.StatusOK, map[string]any{"list": l, "next": cursor}, err)
}
