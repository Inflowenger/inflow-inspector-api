package flowControllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Inflowenger/inflow-inspector-api/env"
	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/contrib/v3/socketio"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"

	"github.com/golang-jwt/jwt/v5"
)

// loadEventHandlersOnce guards against registering the socketio event handlers
// more than once. LoadEventHandlers is invoked from several call sites (route
// setup and the lazy GetWsSessions init); without this guard each call appends
// another set of callbacks, so every event (connect/disconnect/message) would
// fire once per registration.
var loadEventHandlersOnce sync.Once

func LoadEventHandlers() {
	loadEventHandlersOnce.Do(func() {
		socketio.On(socketio.EventConnect, func(ep *socketio.EventPayload) {
			//   ep.Kws.EmitEvent("message", []byte("Welcome!"))
			fmt.Printf("user connected with id : %s\n", ep.Kws.GetStringAttribute("sessId"))

		})
		socketio.On(socketio.EventDisconnect, func(ep *socketio.EventPayload) {
			// Remove the user from the local clients
			fmt.Printf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("sessId"))
			GetWsSessions().Delete(ep.Kws.GetStringAttribute("sessId"))
		})
		socketio.On(socketio.EventMessage, func(ep *socketio.EventPayload) {
			fmt.Printf("Message event - User: %s, Message: %s", ep.Kws.GetStringAttribute("sessId"), string(ep.Data))
			ep.Kws.EmitEvent("message", []byte(fmt.Sprintf("Echo: %s", string(ep.Data))))
		})
	})
}

func wsAuthHandler(c fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		authBody := struct {
			AuthToken string `header:"Authorization" query:"Authorization"`
		}{}
		if err := c.Bind().All(&authBody); err != nil {
			return fiber.ErrBadRequest
		}
		token := authBody.AuthToken
		token = strings.TrimPrefix(token, "Bearer ")
		jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))
		_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwtware.ErrJWTAlg
			}
			return []byte(env.GetJwtSecret()), nil
		})
		if err != nil {
			return fiber.ErrUnauthorized
		}
		return c.Next()
	}
	return fiber.ErrUpgradeRequired

}
func wshandler(kws *socketio.Websocket) {
	sessId := kws.Params("id")

	GetWsSessions().write(sessId, kws.UUID)
	kws.SetAttribute("sessId", sessId)

	welcomeMsg, _ := json.Marshal(fmt.Sprintf("Hello user: %s with UUID: %s", sessId, kws.UUID))
	kws.Emit(welcomeMsg, socketio.TextMessage)

}
func sendmessageto(c fiber.Ctx) error {
	body := struct {
		SessId  string `json:"sessId"`
		Message string `json:"message"`
	}{}
	if err := c.Bind().Body(&body); err != nil {
		return fiber.ErrBadRequest
	}
	err := SendToSession(body.SessId, body.Message)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"status": "message sent"})
}

func SendToSession(sessId string, message string) error {

	if kwsUUID, ok := GetWsSessions().Read(sessId); ok {
		socketio.EmitTo(kwsUUID, []byte(message))
		return nil
	}
	return fmt.Errorf("session ID not found")
}
