package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

// websocket连接
type WebsocketConn struct {
	ws        *websocket.Conn
	mutex     sync.Mutex
	readMutex sync.Mutex
}

// 新建websocket连接
func NewWebsocketConn(ws *websocket.Conn) *WebsocketConn {
	return &WebsocketConn{
		ws:    ws,
		mutex: sync.Mutex{},
	}
}

// 获取websocket连接状态
func (w *WebsocketConn) GetConnStatus() bool {
	return w.ws != nil
}

// 获取websocket连接
func (w *WebsocketConn) GetConn() *websocket.Conn {
	if w.ws == nil {
		return nil
	}

	return w.ws
}

// 获取互斥锁
func (w *WebsocketConn) GetMutex() *sync.Mutex {
	return &w.mutex
}

// 关闭websocket连接
func (w *WebsocketConn) Close() {
	if w.ws == nil {
		return
	}

	w.ws.Close()
}

// 发送文本消息
func (w *WebsocketConn) SendMessage(message string) error {
	if w.ws == nil {
		return errors.New("websocket connection not found")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.ws.WriteMessage(websocket.TextMessage, []byte(message))
}

// 发送二进制消息
func (w *WebsocketConn) SendBinary(bytes []byte) error {
	if w.ws == nil {
		return errors.New("websocket connection not found")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.ws.WriteMessage(websocket.BinaryMessage, bytes)
}

// 发送JSON消息
func (w *WebsocketConn) SendJson(data any) error {
	if w.ws == nil {
		return errors.New("websocket connection not found")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return w.ws.WriteMessage(websocket.TextMessage, jsonData)
}

// 发送文件
func (w *WebsocketConn) SendFile(filepath string) error {
	if w.ws == nil {
		return errors.New("websocket connection not found")
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if err := w.ws.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
			return err
		}
	}

	return nil
}

// 获取消息
func (w *WebsocketConn) GetMessage() (messageType int, message string, err error) {
	if w.ws == nil {
		return 0, "", errors.New("websocket connection not found")
	}

	w.readMutex.Lock()
	defer w.readMutex.Unlock()

	mt, msg, e := w.ws.ReadMessage()
	return mt, string(msg), e
}

// 判断是否是WebSocket关闭错误
func (w *WebsocketConn) IsCloseError(err error) bool {
	closeErr, ok := err.(*websocket.CloseError)

	if ok {
		if closeErr.Code == websocket.CloseNormalClosure ||
			closeErr.Code == websocket.CloseGoingAway ||
			closeErr.Code == websocket.CloseProtocolError ||
			closeErr.Code == websocket.CloseUnsupportedData ||
			closeErr.Code == websocket.CloseNoStatusReceived ||
			closeErr.Code == websocket.CloseAbnormalClosure ||
			closeErr.Code == websocket.CloseInvalidFramePayloadData ||
			closeErr.Code == websocket.ClosePolicyViolation ||
			closeErr.Code == websocket.CloseMessageTooBig ||
			closeErr.Code == websocket.CloseMandatoryExtension ||
			closeErr.Code == websocket.CloseInternalServerErr ||
			closeErr.Code == websocket.CloseServiceRestart ||
			closeErr.Code == websocket.CloseTryAgainLater ||
			closeErr.Code == websocket.CloseTLSHandshake {
			return true
		}
	} else {
		opErr, ok := err.(*net.OpError)
		if ok && opErr.Op == "read" {
			return true
		}
	}

	return false
}

var (
	// 默认websocket配置
	defaultWebsocketUpgrade = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// 将handler转换为websocket handler
func ToWebsocketHandler(handler Handler, websocketUpgrade *websocket.Upgrader) Handler {
	return func(w *Response, r Request) any {
		if r.IsWebsocket() && w.GetWebsocketConn() == nil {
			upgrader := defaultWebsocketUpgrade
			if websocketUpgrade != nil {
				upgrader = websocketUpgrade
			}

			ws, err := upgrader.Upgrade(w.ResponseWriter, r.Request, nil)
			if err != nil {
				return nil
			}
			w.ws = NewWebsocketConn(ws)
		}

		return handler(w, r)
	}
}
