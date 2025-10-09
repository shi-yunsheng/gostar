package handler

import (
	"encoding/json"
	"net/http"
)

// @en response object
//
// @zh 响应对象
type Response struct {
	http.ResponseWriter
	// @en websocket connection
	//
	// @zh websocket连接
	ws *WebsocketConn
	// @en status code
	//
	// @zh 状态码
	StatusCode int
	// @en written
	//
	// @zh 是否已写入
	Written bool
	// @en body
	//
	// @zh 响应体
	body []byte
	// @en break, interrupt the subsequent process, only valid in webapp and static
	//
	// @zh 提前结束，中断后续流程，只在webapp和static中有效
	earlyBreak bool
}

// @en response body
//
// @zh 响应体
type ResponseBody struct {
	Code    uint8  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// @en write header
//
// @zh 写入头
func (w *Response) WriteHeader(code int) {
	if w.Written {
		return
	}
	w.StatusCode = code
	w.Written = true
	w.ResponseWriter.WriteHeader(code)
}

// @en write
//
// @zh 写入
func (w *Response) Write(b []byte) (int, error) {
	if !w.Written {
		w.StatusCode = http.StatusOK
		w.Written = true
	}
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

// @en early break, interrupt the subsequent process, only valid in webapp and static
//
// @zh 提前结束，中断后续流程，只在webapp和static中有效
func (w *Response) EarlyBreak() {
	w.earlyBreak = true
}

// @en get early break, only valid in webapp and static
//
// @zh 获取提前结束状态，只在webapp和static中有效
func (w *Response) GetEarlyBreak() bool {
	return w.earlyBreak
}

// @en set header
//
// @zh 设置响应头
func (w *Response) SetHeader(key string, value string) {
	w.Header().Set(key, value)
}

// @en get header
//
// @zh 获取响应头
func (w *Response) GetHeader(key string) string {
	return w.Header().Get(key)
}

// @en get headers
//
// @zh 获取响应头
func (w *Response) GetHeaders() http.Header {
	return w.Header()
}

// @en get response
//
// @zh 获取响应体
func (w *Response) GetResponse() []byte {
	return w.body
}

// @en set response
//
// @zh 设置响应体
func (w *Response) SetResponse(body []byte) (int, error) {
	w.body = append(w.body, body...)
	return w.ResponseWriter.Write(body)
}

// @en get websocket connection
//
// @zh 获取websocket连接
func (w *Response) GetWebsocketConn() *WebsocketConn {
	return w.ws
}

// @en response json
//
// @zh 响应JSON
func (w *Response) Json(data any) {
	w.SetHeader("Content-Type", "application/json; charset=utf-8")

	jsonData, _ := json.Marshal(data)
	w.Write(jsonData)
}

// @en response html
//
// @zh 响应HTML
func (w *Response) Html(data string) {
	w.SetHeader("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(data))
}

// @en response error
//
// @zh 响应错误
func (w *Response) Error(err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Json(ResponseBody{
		Code:    0,
		Message: err.Error(),
		Data:    nil,
	})
}
