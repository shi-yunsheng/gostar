package handler

import (
	"net/http"
)

// @en convert handler to http.HandlerFunc
//
// @zh 将Handler转换为http.HandlerFunc
func ToHttpHandler(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := &Response{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
			Written:        false,
		}

		request := Request{Request: r}

		handler(response, request)
	}
}
