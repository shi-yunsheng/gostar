package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shi-yunsheng/gostar/router/handler"

	"github.com/shi-yunsheng/gostar/logger"
)

// 限流中间件
func RateLimitMiddleware(requests int, per time.Duration) Middleware {
	// 使用互斥锁保护 map
	var mu sync.Mutex
	// 记录每个 IP+路径 的请求信息
	clients := make(map[string][]time.Time)
	// 清理过期记录
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
		return func(w *handler.Response, r handler.Request) any {
			// 获取客户端 IP 和请求路径
			ip := r.GetClientIP()
			path := r.Request.URL.Path
			// 创建组合键：IP + 路径
			key := strings.Join(ip, ":") + ":" + path

			mu.Lock()
			// 获取该 IP+路径 的请求记录
			times, exists := clients[key]
			if !exists {
				times = []time.Time{}
				clients[key] = times
			}
			// 清除超过限制时间的记录
			cutoff := time.Now().Add(-per)
			var active []time.Time
			for _, t := range times {
				if t.After(cutoff) {
					active = append(active, t)
				}
			}
			clients[key] = active
			// 检查是否超过速率限制
			if len(active) >= requests {
				mu.Unlock()
				logger.W("Rate limit exceeded: IP=%s, Path=%s, Current requests=%d, Limit=%d", ip, path, len(active), requests)
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("Rate limit exceeded"))
				return nil
			}
			// 记录新的请求时间
			clients[key] = append(active, time.Now())
			mu.Unlock()

			return next(w, r)
		}
	}
}
