package chromedpproxy

import (
	"fmt"
	fhwebsocket "github.com/fasthttp/websocket"
	gfwebsocket "github.com/gofiber/websocket/v2"
)

// startWebsocketProxy proxies all messages between Chrome remote debugger websocket to and from requester
func startWebsocketProxy(cdpPort string, requester *gfwebsocket.Conn) (*fhwebsocket.Conn, error) {
	conn, _, err := fhwebsocket.DefaultDialer.Dial(fmt.Sprintf("ws://127.0.0.1:%s/devtools/page/%s", cdpPort, requester.Params("id")), nil)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			err = requester.WriteMessage(messageType, message)
			if err != nil {
				break
			}
		}
		_ = conn.Close()
	}()
	return conn, nil
}
