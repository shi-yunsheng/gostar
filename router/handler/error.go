package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/shi-yunsheng/gostar/utils"
)

var debug = false

// @en enable debug mode
//
// @zh 开启调试模式
func EnableDebug() {
	debug = true
}

type ErrorHtml struct {
	Title       string `json:"title"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Message     string `json:"message"`
	Stack       string `json:"stack"`
}

var errorTemplate = `<!DOCTYPE html>
<html lang="">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}

		html {
			line-height: 1.15;
			-webkit-text-size-adjust: 100%;
		}

		body {
			margin: 0;
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif;
			font-size: 14px;
			line-height: 1.5;
		}

		:root {
			--scrollbar-color: #c1c1c1;
			--scrollbar-background: #ececec;
			--scrollbar-track: #f2f3f5;
			--text-color-1: #000000;
			--text-color-2: #181818;
			--background-color-1: #ffffff;
			--background-color-2: #f2f3f5;
		}

		@media (prefers-color-scheme: dark) {
			:root {
				--scrollbar-color: #444444;
				--scrollbar-background: #222222;
				--scrollbar-track: #181a1b;
				--text-color-1: #f5f5f5;
				--text-color-2: #cccccc;
				--background-color-1: #181a1b;
				--background-color-2: #23272e;
			}
		}

		::-webkit-scrollbar {
			width: 0.5rem;
			background: var(--scrollbar-background);
		}

		::-webkit-scrollbar-thumb {
			background: var(--scrollbar-color);
			border-radius: 0.25rem;
		}

		::-webkit-scrollbar-track {
			background: var(--scrollbar-track);
		}
			
        .container {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100svh;
            margin: 0;
            gap: 0.1em;
			color: var(--text-color-1);
			background-color: var(--background-color-1);
			mix-blend-mode: difference;
        }

        .code {
            font-size: 6em;
            font-weight: 500;
            letter-spacing: 0.05em;
        }

        .description {
            font-size: 1.3em;
            font-weight: 500;
            letter-spacing: 0.05em;
        }
        
        .message {
            font-size: 1em;
            font-weight: 400;
            letter-spacing: 0.05em;
        }

		.stack {
			font-size: 0.9em;
			color: var(--text-color-2);
			background-color: var(--background-color-2);
			padding: 1rem;
			border-radius: 0.625rem;
			margin-top: 0.5rem;
			font-weight: 400;
			letter-spacing: 0.05em;
			white-space: pre-wrap;
			word-break: break-all;
			word-wrap: break-word;
			overflow-wrap: break-word;
			overflow: auto;
    		max-height: 70vh;
		}
    </style>
</head>
<body>
    <div class="container">
        <div class="code">{{.Code}}</div>
        <div class="description">{{.Description}}</div>
        <div class="message">{{.Message}}</div>
		{{if .Stack}}
		<div class="stack">{{.Stack}}</div>
		{{end}}
    </div>
</body>
</html>`

// @en get error html
// @zh 获取错误html
func errorPage(errorHtml ErrorHtml) string {
	tmpl := template.Must(template.New("error").Parse(errorTemplate))
	var buf bytes.Buffer
	tmpl.Execute(&buf, errorHtml)

	return buf.String()
}

// @en 404 not found
//
// @zh 404 页面不存在
func NotFound(w *Response, r Request) {
	w.WriteHeader(http.StatusNotFound)

	errorHtml := ErrorHtml{
		Title:       "404 Not Found",
		Code:        "404",
		Description: "Not Found",
		Message:     "Sorry, the page you visited does not exist.",
	}

	if debug {
		errorHtml.Stack = utils.GetStackTrace()
	}

	w.Html(errorPage(errorHtml))
}

// @en 405 method not allowed
//
// @zh 405 请求方法不允许
func MethodNotAllowed(w *Response, r Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)

	errorHtml := ErrorHtml{
		Title:       "405 Method Not Allowed",
		Code:        "405",
		Description: "Method Not Allowed",
		Message:     "Sorry, the method you used is not allowed.",
	}

	if debug {
		errorHtml.Stack = utils.GetStackTrace()
	}

	w.Html(errorPage(errorHtml))
}

// @en 401 unauthorized
//
// @zh 401 未授权
func Unauthorized(w *Response, r Request) {
	w.WriteHeader(http.StatusUnauthorized)

	errorHtml := ErrorHtml{
		Title:       "401 Unauthorized",
		Code:        "401",
		Description: "Unauthorized",
		Message:     "Sorry, you are not authorized to access this page.",
	}

	if debug {
		errorHtml.Stack = utils.GetStackTrace()
	}

	w.Html(errorPage(errorHtml))
}

// @en 403 forbidden
//
// @zh 403 禁止访问
func Forbidden(w *Response, r Request) {
	w.WriteHeader(http.StatusForbidden)

	errorHtml := ErrorHtml{
		Title:       "403 Forbidden",
		Code:        "403",
		Description: "Forbidden",
		Message:     "Sorry, you are not allowed to access this page.",
	}

	if debug {
		errorHtml.Stack = utils.GetStackTrace()
	}

	w.Html(errorPage(errorHtml))
}

// @en 500 internal server error
//
// @zh 500 内部服务器错误
func InternalServerError(w *Response, r Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)

	errorHtml := ErrorHtml{
		Title:       "500 Internal Server Error",
		Code:        "500",
		Description: "Internal Server Error",
		Message:     fmt.Sprintf("Sorry, the server is busy, please try again later. ERROR: %s", err.Error()),
	}

	if debug {
		errorHtml.Stack = utils.GetStackTrace()
	}

	w.Html(errorPage(errorHtml))
}

// @zh 400 bad request
//
// @zh 400 请求错误
func BadRequest(w *Response, r Request, err error) {
	w.WriteHeader(http.StatusBadRequest)

	errorHtml := ErrorHtml{
		Title:       "400 Bad Request",
		Code:        "400",
		Description: "Bad Request",
		Message:     fmt.Sprintf("Sorry, the request is invalid. ERROR: %s", err.Error()),
	}

	if debug {
		errorHtml.Stack = utils.GetStackTrace()
	}

	w.Html(errorPage(errorHtml))
}
