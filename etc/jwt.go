package etc

import (
	"github.com/Inflowenger/inflow-inspector-api/env"
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
)

func HS256SecKeyHandler() fiber.Handler {

	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(env.GetJwtSecret())},
	})
}
