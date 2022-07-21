// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package controller

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"nautilus/golib/log"
)

var (
	BUFFER_SIZE = 1024
	PING        = []byte("PING")
	QUIT        = []byte("quit")
	FINISH      = []byte("finish")
)

func NewWebsocket() *WSocket {
	return &WSocket{}
}

type WSocket struct {
	conn *websocket.Conn
}

func (w *WSocket) Serve(c *gin.Context) error {
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
		log.Errorf("http request upgrade to websocket err: %s", err)
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
		log.Infof("websocket closed")
		return err
	})
	w.conn = conn
	log.Infof("create websocket success")
	return nil
}

func (w *WSocket) Heartbeat() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("---- websocket read heartbeat exception: %s", err)
		}
	}()
	for {
		code, data, err := w.conn.ReadMessage()
		if err != nil {
			log.Errorf("read message error: %s", err)
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

func (w *WSocket) Echo(msg string) {
	w.Send([]byte(msg))
}

func (w *WSocket) Send(msg []byte) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("---- websocket send message exception: %s", err)
		}
	}()
	w.conn.WriteMessage(websocket.TextMessage, msg)
}

func (w *WSocket) Quit() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("---- websocket send quit exception: %s", err)
		}
	}()
	w.conn.WriteMessage(websocket.TextMessage, QUIT)
}

func (w *WSocket) Finish() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("---- websocket send finish exception: %s", err)
		}
	}()
	w.conn.WriteMessage(websocket.TextMessage, FINISH)
}

// Realtime 执行命令的实时输出
func (w *WSocket) Realtime(param string, output *string) {
	cmd := exec.Command("bash", "-c", param)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		w.Echo(err.Error())
		w.Quit()
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		w.Echo(err.Error())
		w.Quit()
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go w.read(&wg, stdout, output)
	go w.read(&wg, stderr, output)

	if err := cmd.Start(); err != nil {
		w.Echo(err.Error())
		w.Quit()
		return
	}

	if err := cmd.Wait(); err != nil {
		w.Echo(err.Error())
		w.Quit()
		return
	}

	if !cmd.ProcessState.Success() {
		w.Quit()
		return
	}
}

func (w *WSocket) read(wg *sync.WaitGroup, std io.ReadCloser, output *string) {
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
