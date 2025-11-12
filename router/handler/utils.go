package handler

import (
	"net/http"
)

// 将Handler转换为http.HandlerFunc
func ToHttpHandler(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &Response{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
			Written:        false,
		}

		request := Request{Request: r}

		resp := handler(response, request)
		// 如果响应已经写入，则不处理返回值
		if response.Written {
			return
		}
		// 返回包装在ResponseBody中的数据
		response.Json(resp)
	}
}
