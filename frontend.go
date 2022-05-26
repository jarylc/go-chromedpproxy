package chromedpproxy

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

// avoid any packaging by including the front-end html as a variable
const html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta charset="UTF-8">
    <title>Remote Chrome Frontend</title>
    <style>
        canvas {
            width: 90vw;
            height: 90vh;
            border: gray dotted 1px;
            padding: 0;
            margin: auto;
            display: block;
        }
        input {
            width: 90vw;
            padding: 0;
            margin: auto;
            display: block;
            font-size: 1em;
        }
    </style>
</head>
<body>
    <canvas tabindex="1" oncontextmenu="return false;"></canvas>
    <input id="insert" type="text" placeholder="Enter text to paste (press ENTER key to submit)"/>
</body>
<footer>
    <script>
      const urlParams = new URLSearchParams(window.location.search);
      const target = urlParams.get('id') || "";

      const canvas = document.querySelector("canvas");
      const ctx = canvas.getContext("2d");
      const insert = document.getElementById("insert");

      let sessionId = null;

      let _id = 0;
      const id = () => {
        return _id++;
      };

      const ws = new WebSocket(((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/ws/" + target);
      ws.onopen = function() {
        ws.send(JSON.stringify({
          id: id(),
          method: 'Target.attachToTarget',
          params: {
            targetId: target,
            flatten: true,
          },
        }));
      };
      ws.onmessage = function(e) {
        const data = JSON.parse(e.data)
        if ('error' in data && data.error.message === 'No target with given id found') {
          ws.close()
        }
        switch (data.method) {
          case "Target.attachedToTarget":
            sessionId = data.params.sessionId
            ws.send(JSON.stringify({
              sessionId,
              id: id(),
              method: 'Page.startScreencast',
              params: {
                format: "jpeg",
                quality: 100
              },
            }));
            resizeEvent()
            break;
          case "Page.screencastFrame":
            let image = data.params.data;
            let img = new Image();

            img.onload = function() {
              ctx.drawImage(img, 0, 0);
            };
            img.src = "data:image/jpeg;base64," + image;
            ws.send(JSON.stringify({
              sessionId,
              id: id(),
              method: 'Page.screencastFrameAck',
              params: {
                sessionId: data.params.sessionId
              },
            }));
            break;
          default:
            break;
        }
      };
      ws.onclose = function(e) {
        const body = document.querySelector("body");
        if (sessionId === null) {
          body.innerText = "No session was found";
        } else {
          body.innerText = "Connection has been closed";
        }
      };

      const mouseEvent = (e) => {
        const buttons = { 1: 'left', 2: 'middle', 3: 'right' };
        const event = e.type === 'wheel' ? window.event || e : e;
        const types = {
          mousedown: 'mousePressed',
          mouseup: 'mouseReleased',
          wheel: 'mouseWheel',
          touchstart: 'mousePressed',
          touchend: 'mouseReleased',
        };

        if (!(e.type in types) || (event.type !== 'wheel' && event.which === 0 && event.type !== 'mousemove')) {
          return;
        }

        const type = types[event.type];
        const isScroll = type.indexOf('wheel') !== -1;
        const x = isScroll ? event.clientX : event.offsetX;
        const y = isScroll ? event.clientY : event.offsetY;

        const params = {
          type: types[event.type],
          x,
          y,
          button: event.type === 'wheel' ? 'none' : buttons[event.which],
          clickCount: 1,
        };
        if (event.type === 'wheel') {
          params.deltaX = event.wheelDeltaX || 0;
          params.deltaY = -event.wheelDeltaY || -event.wheelDelta;
        }

        ws.send(JSON.stringify({
          sessionId,
          id: id(),
          method: 'Input.dispatchMouseEvent',
          params,
        }));
      }
      canvas.addEventListener("mousedown", mouseEvent, true)
      canvas.addEventListener("mouseup", mouseEvent, true)
      canvas.addEventListener("wheel", mouseEvent, true)
      canvas.addEventListener("touchstart", mouseEvent, true)
      canvas.addEventListener("touchend", mouseEvent, true)

      const keyEvent = (e) => {
        if (e.keyCode === 8) {
          e.preventDefault();
        }

        let type;
        switch (e.type) {
          case 'keydown':
            type = 'keyDown';
            break;
          case 'keyup':
            type = 'keyUp';
            break;
          case 'keypress':
            type = 'char';
            break;
          default:
            return;
        }

        const text = type === 'char' ? String.fromCharCode(e.charCode) : undefined;
        ws.send(JSON.stringify({
          sessionId,
          id: id(),
          method: 'Input.dispatchKeyEvent',
          params: {
            type,
            text,
            unmodifiedText: text ? text.toLowerCase() : undefined,
            keyIdentifier: e.keyIdentifier,
            code: e.code,
            key: e.key,
            windowsVirtualKeyCode: e.keyCode,
            nativeVirtualKeyCode: e.keyCode,
            autoRepeat: false,
            isKeypad: false,
            isSystemKey: false,
          },
        }));
      };
      canvas.addEventListener("keydown", keyEvent, true)
      canvas.addEventListener("keyup", keyEvent, true)
      canvas.addEventListener("keypress", keyEvent, true)

      insert.addEventListener('keydown', (e) => {
        if (e.keyCode === 13) {
          const text = e.target.value;
          ws.send(JSON.stringify({
            sessionId,
            id: id(),
            method: 'Input.insertText',
            params: {
              text
            },
          }));
          e.target.value = '';
        }
      }, false);

      const resizeEvent = () => {
        if (sessionId === null)
          return

        let { width, height } = canvas.getBoundingClientRect();
        width = Math.floor(width);
        height = Math.floor(height);

        ctx.canvas.width = width;
        ctx.canvas.height = height;

        ws.send(JSON.stringify({
          sessionId,
          id: id(),
          method: 'Emulation.setDeviceMetricsOverride',
          params: {
            width,
            height,
            deviceScaleFactor: 1,
            mobile: (width <= 760)
          },
        }));
      };
      window.addEventListener('resize', resizeEvent, false);
    </script>
</footer>
</html>`

var timeout = time.Second * 15

// startFrontEnd starts a blocking web server that serves the front-end alongside the websocket proxy
func startFrontEnd(frontendListenAddr string, cdpPort string, cancelChan chan bool) {
	r := mux.NewRouter()

	srv := &http.Server{
		Addr:         frontendListenAddr,
		WriteTimeout: timeout,
		ReadTimeout:  timeout,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(html))
		if err != nil {
			log.Printf("Error writing response: %s", err)
		}
	})
	r.HandleFunc("/ws/{id}", func(writer http.ResponseWriter, request *http.Request) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Printf("Error upgrading connection: %s", err)
			return
		}

		id := mux.Vars(request)["id"]
		proxy, err := startWebsocketProxy(cdpPort, id, conn)
		if err != nil {
			log.Printf("Error starting websocket proxy: %s", err)
			return
		}
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			err = proxy.WriteMessage(messageType, message)
			if err != nil {
				break
			}
		}
		_ = proxy.Close()
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Printf("Error starting frontend: %s", err)
			}
		}
	}()

	<-cancelChan
	stopFrontEnd(srv)
}

func stopFrontEnd(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		log.Printf("Error shutting down frontend: %v", err)
		return
	}
}
