package handler

import (
	"encoding/json"
	"net/http"
)

// 响应对象
type Response struct {
	http.ResponseWriter
	// websocket连接
	ws *WebsocketConn
	// 状态码
	StatusCode int
	// 是否已写入
	Written bool
	// 响应体
	body []byte
	// 提前结束，中断后续流程，只在webapp和static中有效
	earlyBreak bool
}

// 响应体
type ResponseBody struct {
	Code    int    `json:"code"`
	Show    bool   `json:"show"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// 写入头
func (w *Response) WriteHeader(code int) {
	if w.Written {
		return
	}
	w.StatusCode = code
	w.Written = true
	w.ResponseWriter.WriteHeader(code)
}

// 写入
func (w *Response) Write(b []byte) (int, error) {
	if !w.Written {
		w.StatusCode = http.StatusOK
		w.Written = true
	}
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

// 提前结束，中断后续流程，只在webapp和static中有效
func (w *Response) EarlyBreak() {
	w.earlyBreak = true
}

// 获取提前结束状态，只在webapp和static中有效
func (w *Response) GetEarlyBreak() bool {
	return w.earlyBreak
}

// 设置响应头
func (w *Response) SetHeader(key string, value string) {
	w.Header().Set(key, value)
}

// 获取响应头
func (w *Response) GetHeader(key string) string {
	return w.Header().Get(key)
}

// 获取响应头
func (w *Response) GetHeaders() http.Header {
	return w.Header()
}

// 获取响应体
func (w *Response) GetResponse() []byte {
	return w.body
}

// 设置响应体
func (w *Response) SetResponse(body []byte) (int, error) {
	w.body = append(w.body, body...)
	return w.ResponseWriter.Write(body)
}

// 设置Cookie
func (w *Response) SetCookie(key, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// 删除Cookie
func (w *Response) DeleteCookie(key string) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// 获取websocket连接
func (w *Response) GetWebsocketConn() *WebsocketConn {
	return w.ws
}

// 响应JSON
func (w *Response) Json(data any) {
	w.SetHeader("Content-Type", "application/json; charset=utf-8")

	jsonData, _ := json.Marshal(data)
	w.Write(jsonData)
}

// 响应HTML
func (w *Response) Html(data string) {
	w.SetHeader("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(data))
}

// 响应Text
func (w *Response) Text(data string) {
	w.SetHeader("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(data))
}
