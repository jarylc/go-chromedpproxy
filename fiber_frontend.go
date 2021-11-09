package chromedpproxy

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
	"os"
	"os/signal"
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

// startFrontEnd starts a blocking Fiber web server that serves the front-end alongside the websocket proxy
func startFrontEnd(frontendListenAddr string, cdpPort string, cancelChan chan bool) {
	app := fiber.New(fiber.Config{
		ReduceMemoryUsage:     true,
		DisableStartupMessage: true,
	})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		select {
		case <-interrupt:
			app.Shutdown()
		case <-cancelChan:
			app.Shutdown()
		}
	}()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return c.SendString(html)
	})
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/:id", websocket.New(func(conn *websocket.Conn) {
		proxy, err := startWebsocketProxy(cdpPort, conn)
		if err != nil {
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
	}))
	err := app.Listen(frontendListenAddr)
	if err != nil {
		log.Panic(err)
		return
	}
}
