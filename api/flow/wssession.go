package flowControllers

import (
	"fmt"
	"sync"

	fuse "github.com/Inflowenger/inflow-fusion/inflow"
	"github.com/gofiber/contrib/v3/socketio"
	"github.com/nats-io/nats.go"
)

var sessManager *wsSessions



type wsSessions struct {
	sess  map[string]string
	mutex sync.RWMutex
}

func (d *wsSessions) Read(key string) (string, bool) {
	d.mutex.RLock()

	defer d.mutex.RUnlock()
	val, exists := d.sess[key]
	return val, exists
}

func (d *wsSessions) write(key string, value string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.sess[key] = value
}

func (d *wsSessions) Delete(key string) {
	d.mutex.RLock()

	defer d.mutex.RUnlock()
	delete(d.sess, key)
}


func GetWsSessions() *wsSessions {
	if sessManager == nil {
		sessManager = &wsSessions{sess: make(map[string]string)}
		LoadEventHandlers()
		sendInflowLogs()
	}
	return sessManager
}

func sendInflowLogs() {
	natsConn,err:=fuse.GetInflowBackend().GetInflowEventsPipe()
	if err != nil {
		fmt.Printf("Error getting NATS connection: %v\n", err)
		return
	}
	natsConn.Subscribe("inflow.event.log", func(msg *nats.Msg) {
		logMessage := string(msg.Data)
		// fmt.Printf("Received log message: %s\n", logMessage)
		// Broadcast the log message to all connected WebSocket clients
		socketio.Broadcast([]byte(logMessage), socketio.TextMessage)
	})
}