package handler

// 处理器函数
type Handler func(w *Response, r Request) any
