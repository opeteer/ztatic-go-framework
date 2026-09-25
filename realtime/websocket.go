package realtime

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev/API flexibility
	},
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// WebSocketHandler upgrades the HTTP connection to a full-duplex WebSocket stream.
// It maps the topic subscription similarly to the SSE Handler, but allows bidirectional 
// binary/text frames for advanced use cases.
func WebSocketHandler(broker EventBroker) echo.HandlerFunc {
	return func(c *echo.Context) error {
		topic := c.QueryParam("topic")
		if topic == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Missing 'topic' query parameter")
		}

		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		ctx := c.Request().Context()
		stream, unsubscribe := broker.Subscribe(ctx, topic)

		defer func() {
			unsubscribe()
			ws.Close()
		}()

		ws.SetReadLimit(maxMessageSize)
		ws.SetReadDeadline(time.Now().Add(pongWait))
		ws.SetPongHandler(func(string) error {
			ws.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

		// Read pump to process pong messages and detect disconnects
		go func() {
			defer ws.Close()
			for {
				_, _, err := ws.ReadMessage()
				if err != nil {
					break
				}
			}
		}()

		pingTicker := time.NewTicker(pingPeriod)
		defer pingTicker.Stop()

		for {
			select {
			case msg, ok := <-stream:
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					ws.WriteMessage(websocket.CloseMessage, []byte{})
					return nil
				}
				if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					return nil
				}
			case <-pingTicker.C:
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
					return nil
				}
			case <-ctx.Done():
				return nil
			}
		}
	}
}
