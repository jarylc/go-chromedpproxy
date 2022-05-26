package chromedpproxy

import (
	"fmt"
	"github.com/gorilla/websocket"
)

// startWebsocketProxy proxies all messages between Chrome remote debugger websocket to and from requester
func startWebsocketProxy(cdpPort string, id string, requester *websocket.Conn) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://127.0.0.1:%s/devtools/page/%s", cdpPort, id), nil)
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
