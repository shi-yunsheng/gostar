package middleware

import (
	"gostar/router/handler"
	"net/http"
	"strings"
	"sync"
	"time"

	"gostar/logger"
)

// @en rate limit middleware
//
// @zh 限流中间件
func RateLimitMiddleware(requests int, per time.Duration) Middleware {
	// @en use mutex to protect map
	// @zh 使用互斥锁保护 map
	var mu sync.Mutex

	// @en record request information for each IP+path
	// @zh 记录每个 IP+路径 的请求信息
	clients := make(map[string][]time.Time)

	// @en clean expired records
	// @zh 清理过期记录
	go func() {
		for {
			time.Sleep(per)
			mu.Lock()
			for key, times := range clients {
				var active []time.Time
				cutoff := time.Now().Add(-per)
				for _, t := range times {
					if t.After(cutoff) {
						active = append(active, t)
					}
				}

				if len(active) == 0 {
					delete(clients, key)
				} else {
					clients[key] = active
				}
			}
			mu.Unlock()
		}
	}()

	return func(next handler.Handler) handler.Handler {
		return func(w *handler.Response, r handler.Request) {
			// @en get client IP and request path
			// @zh 获取客户端 IP 和请求路径
			ip := r.GetClientIP()
			path := r.Request.URL.Path

			// @en create combined key: IP + path
			// @zh 创建组合键：IP + 路径
			key := strings.Join(ip, ":") + ":" + path

			mu.Lock()

			// @en get request record for IP+path
			// @zh 获取该 IP+路径 的请求记录
			times, exists := clients[key]
			if !exists {
				times = []time.Time{}
				clients[key] = times
			}

			// @en clean expired records
			// @zh 清除超过限制时间的记录
			cutoff := time.Now().Add(-per)
			var active []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					active = append(active, t)
				}
			}
			clients[key] = active

			// @en check if exceeds rate limit
			// @zh 检查是否超过速率限制
			if len(active) >= requests {
				mu.Unlock()
				logger.W("Rate limit exceeded: IP=%s, Path=%s, Current requests=%d, Limit=%d", ip, path, len(active), requests)
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Rate limit exceeded"))
				return
			}

			// @en record new request time
			// @zh 记录新的请求时间
			clients[key] = append(active, time.Now())
			mu.Unlock()

			next(w, r)
		}
	}
}
