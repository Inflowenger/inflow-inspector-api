package main

import (
	"fmt"

	apiHandlers "github.com/Inflowenger/inflow-inspector-api/api"
	"github.com/Inflowenger/inflow-inspector-api/env"
	"github.com/Inflowenger/inflow-inspector-api/inflow"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	err := inflow.InitInflowConnection()
	if err != nil {
		panic(err)
	}
	err = inflow.LoadSvcNodehandlers()
	if err != nil {
		panic(err)
	}
	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
		AppName:     "inflow-developer",
	})

	app.Use(cors.New())

	app.Use(logger.New())
	apiHandlers.RegisterAll(app)
	if err := app.Listen(env.GetApiPort()); err != nil {
		panic(err)
	}
}
