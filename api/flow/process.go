package flowControllers

import (
	"fmt"

	inflowfuse "github.com/Inflowenger/inflow-fusion/inflow"
	"github.com/Inflowenger/inflow-inspector-api/etc"
	"github.com/Inflowenger/inflow-inspector-api/inflow"
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/Inflowenger/inflow-inspector-api/repository"
	"github.com/gofiber/fiber/v3"
)

func newProcess(c fiber.Ctx) error {
	input := models.ProcessRequestInput{}
	if err := c.Bind().Body(&input); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	rec, err := repository.GetFlowById(input.FlowId)
	if err != nil {
		return etc.Send(c, fiber.StatusNotFound, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: "given flow id not found or error  occured " + err.Error()})
	}
	startNodeId, err := inflow.GetStartNodeId(*rec)
	if err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Message: err.Error()})
	}

	proc, err := inflowfuse.NewProcess(startNodeId,
		inflowfuse.WithContextDocument(input.ContextId),
		inflowfuse.WithFlowId(input.FlowId),
		// inflowfuse.WithInflowToken("http://mate-Predator-PHN16-73:9001", ""), //empty token means use infra token
		// inflowfuse.WithInflowJwtSecret("http://mate-Predator-PHN16-73:9001","l9hJN8YbNf4tTNqC2Nu221ld5DLjFhgzCAxU4pydTUNkKlscd0F0VlVDGt1d1A1B"),
		inflowfuse.WithMeta(map[string]string{"account": "dev"}),
	)
	if err != nil {
		fmt.Println(err.Error())

	}
	resp, err := proc.Exec(c.Context())
	if err != nil {
		return etc.Send(c, fiber.StatusNotAcceptable, nil, map[string]any{"response": resp, "error": err.Error()})

	}
	return etc.Send(c, fiber.StatusAccepted, map[string]any{"pid": resp.Data.PID, "selected_resource": proc.GetResource()}, err)
}
func compile(c fiber.Ctx) error {
	input := models.ProcessRequestInput{}
	if err := c.Bind().Body(&input); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	rec, err := repository.GetFlowById(input.FlowId)
	if err != nil {
		return etc.Send(c, fiber.StatusNotFound, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: "given flow id not found or error  occured " + err.Error()})
	}
	startNodeId, cmp, err := inflow.FLowCompiler(*rec)
	if err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Message: err.Error()})
	}

	proc, err := inflowfuse.NewProcess(startNodeId,
		inflowfuse.WithContextDocument(input.ContextId),
		inflowfuse.WithFlowId(input.FlowId),
		inflowfuse.WithInflowToken("http://mate-Predator-PHN16-73:9001", ""),
		inflowfuse.WithMeta(map[string]string{"account": "dev"}),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	return etc.Send(c, fiber.StatusAccepted, map[string]any{"selected_resource": proc.GetResource(), "process_req": proc.GetRequest(), "compiled": cmp}, err)
}
func stopByPid(c fiber.Ctx) error {
	input := models.StopRequest{}
	if err := c.Bind().Body(&input); err != nil {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: fiber.ErrBadRequest.Message})
	}
	pid := c.Params("pid")
	if pid == "" {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: "pid is required"})
	}
	if input.Resource == "" {
		return etc.Send(c, fiber.StatusBadRequest, nil, models.ErrorResponse{Code: fiber.ErrBadRequest.Code, Message: "resource is required"})
	}
	ps, err := inflowfuse.StopProcess(c.Context(), pid, input.Resource)

	if err != nil {
		return etc.Send(c, fiber.StatusNotAcceptable, nil, map[string]any{"error": err.Error()})
	}

	return etc.Send(c, fiber.StatusAccepted, map[string]any{"response": ps}, nil)
}
