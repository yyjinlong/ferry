// copyright @ 2021 ops inc.
//
// author: jinlong yang
//

package ws

import (
	"bytes"
	"ferry/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
