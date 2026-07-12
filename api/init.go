package apiHandlers

import (
	"fmt"
	"strings"

	contextControllers "github.com/Inflowenger/inflow-inspector-api/api/context"
	extensionControllers "github.com/Inflowenger/inflow-inspector-api/api/extension"
	flowControllers "github.com/Inflowenger/inflow-inspector-api/api/flow"
	"github.com/Inflowenger/inflow-inspector-api/env"
	"github.com/Inflowenger/inflow-inspector-api/etc"
	"github.com/Inflowenger/inflow-inspector-api/models"
	"github.com/go-playground/validator/v10"
	vf "github.com/mehdi-shokohi/fiberValidation"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/proxy"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
)

func RegisterAll(api fiber.Router) {
	validatorLoad()
	api.Use(recoverer.New(recoverer.Config{PanicHandler: func(c fiber.Ctx, r any) error {
		return c.Status(fiber.StatusBadRequest).JSON(models.Response{Data: nil, Error: r})
	}}))
	api.All("/infra/*", etc.HS256SecKeyHandler(), infraProxyHandler)
	flowControllers.Register(api)
	contextControllers.Register(api)
	extensionControllers.Register(api)
}

func infraProxyHandler(c fiber.Ctx) error {
	url := fmt.Sprintf("%s/%s?%s", env.GetInfraApiUrl(), c.Params("*1"), c.Request().URI().QueryString())
	return proxy.Forward(url)(c)
}

// load validation rules
func validatorLoad() {
	vfi := vf.NewFiberValidation(vf.WithResponseCast(func(errs []vf.ValidationError) any {
		errList := []map[string]any{}
		for _, el := range errs {
			errList = append(errList, map[string]any{
				"field":   el.Field,
				"ns":      el.NameSpace,
				"message": el.Message,
			})
		}
		return models.Response{Data: nil, Error: errList}
	}))
	vfi.RegisterValidation("r_required", func(fl validator.FieldLevel) bool {
		if len(strings.TrimSpace(fl.Field().String())) > 0 {
			return true
		}
		return false
	}, "{0} field is required", func(fe validator.FieldError) []string {
		return []string{fe.Field()}
	})
}
