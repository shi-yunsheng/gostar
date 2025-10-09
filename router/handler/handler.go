package handler

// @en handler function
//
// @zh 处理器函数
type Handler func(w *Response, r Request)
