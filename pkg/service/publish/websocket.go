// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package publish

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var (
	BUFFER_SIZE = 1024
	PING        = []byte("PING")
	QUIT        = []byte("quit")
	FINISH      = []byte("finish")
)

type WebSocket struct {
	conn *websocket.Conn
}

func NewWebsocket() *WebSocket {
	return &WebSocket{}
}

func (w *WebSocket) Serve(c *gin.Context) error {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  BUFFER_SIZE,
		WriteBufferSize: BUFFER_SIZE,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// 将当前请求升级为websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("upgrade http request to websocket err: %s", err)
		return err
	}

	// websocket协议对应的ping和pong回调方法
	conn.SetPingHandler(func(s string) error {
		return conn.WriteMessage(websocket.PingMessage, []byte(s))
	})
	conn.SetPongHandler(func(s string) error {
		return conn.WriteMessage(websocket.PongMessage, []byte(s))
	})

	// websocket协议对应的close回调方法
	conn.SetCloseHandler(func(code int, text string) error {
		log.Infof("upgrade http request failed: websocket closed")
		return err
	})
	w.conn = conn
	log.Infof("upgrade http request to websocket success")
	return nil
}

func (w *WebSocket) Heartbeat() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("websocket read heartbeat exception: %s", err)
		}
	}()

	for {
		code, data, err := w.conn.ReadMessage()
		if err != nil {
			log.Errorf("websocket read message error: %s", err)
			return
		}
		if code == -1 {
			return
		}
		if bytes.Compare(data, PING) == 0 {
			continue
		}
	}
}

func (w *WebSocket) Echo(msg string) {
	w.Send([]byte(msg))
}

func (w *WebSocket) EchoLine(msg string) {
	w.Send([]byte(msg + "\n"))
}

func (w *WebSocket) EchoRed(msg string) {
	w.Send([]byte(fmt.Sprintf("\033[31m%s\033[0m\n", msg)))
}

func (w *WebSocket) EchoGreen(msg string) {
	w.Send([]byte(fmt.Sprintf("\033[32m%s\033[0m\n", msg)))
}

func (w *WebSocket) Send(msg []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("websocket send message exception: %s", err)
		}
	}()
	w.conn.WriteMessage(websocket.TextMessage, msg)
}

func (w *WebSocket) Quit() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("websocket send quit exception: %s", err)
		}
	}()
	w.conn.WriteMessage(websocket.TextMessage, QUIT)
}

func (w *WebSocket) Finish() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("websocket send finish exception: %s", err)
		}
	}()
	w.conn.WriteMessage(websocket.TextMessage, FINISH)
}

// Realtime 执行命令的实时输出
func (w *WebSocket) Realtime(param string, output *string) error {
	cmd := exec.Command("bash", "-c", param)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		w.Echo(err.Error())
		w.Quit()
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		w.Echo(err.Error())
		w.Quit()
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go w.read(&wg, stdout, output)
	go w.read(&wg, stderr, output)

	if err := cmd.Start(); err != nil {
		w.Echo(err.Error())
		w.Quit()
		return err
	}

	if err := cmd.Wait(); err != nil {
		w.Echo(err.Error())
		w.Quit()
		return err
	}

	if !cmd.ProcessState.Success() {
		w.Quit()
		return err
	}
	return nil
}

func (w *WebSocket) read(wg *sync.WaitGroup, std io.ReadCloser, output *string) {
	defer wg.Done()

	reader := bufio.NewReader(std)
	for {
		buf, err := reader.ReadString('\n')
		if err != nil || err == io.EOF {
			return
		}
		w.Echo(buf)
		*output += strings.Replace(buf, "\u0000", "", -1)
	}
}
