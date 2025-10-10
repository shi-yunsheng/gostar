package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strings"

	"github.com/shi-yunsheng/gostar/utils"
)

// @en request object
//
// @zh 请求对象
type Request struct {
	*http.Request
	params []Param
	model  any
}

// @en set parameters
//
// @zh 设置参数
func (r *Request) SetParams(params []Param) {
	r.params = params
}

// @en get parameter
//
// @zh 获取参数
func (r *Request) GetParam(key string) any {
	for _, param := range r.params {
		if param.Key == key {
			return param.Value
		}
	}

	return nil
}

// @en get binding parameter model
//
// @zh 获取绑定参数模型
func (r *Request) GetBindModel() any {
	return r.model
}

// @en set binding parameter model
//
// @zh 设置绑定参数模型
func (r *Request) SetBindModel(model any) {
	r.model = model
}

// @en get query
//
// @zh 获取查询参数
func (r *Request) GetQuery(key string, defaultVal ...any) any {
	query := r.GetAllQuery()
	if val, ok := query[key]; ok {
		return val
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return nil
}

// @en get all query parameters
//
// @zh 获取所有查询参数
func (r *Request) GetAllQuery() map[string]any {
	query := r.URL.Query()
	result := make(map[string]any)
	for key, values := range query {
		if len(values) == 1 {
			result[key] = values[0]
		} else {
			result[key] = values
		}
	}
	return result
}

// @en get body
//
// @zh 获取请求体
func (r *Request) GetBody(key string) (any, error) {
	body, err := r.GetAllBody()
	if err != nil {
		return nil, err
	}
	if val, ok := body[key]; ok {
		return val, nil
	}
	return nil, errors.New("key not found")
}

// @en get all body data
//
// @zh 获取所有请求体数据
func (r *Request) GetAllBody() (map[string]any, error) {
	if r.Body == nil {
		return nil, errors.New("request body is empty")
	}

	// @en read request body only once
	// @zh 只读取一次请求体
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}

	// @en put the content back for subsequent reads
	// @zh 将读取的内容放回，以便后续再次读取
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var jsonData map[string]any
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return nil, err
	}

	return jsonData, nil
}

// @en get client IP
//
// @zh 获取客户端IP
func (r *Request) GetClientIP() []string {
	var ips []string

	// @en check X-Forwarded-For header first
	// @zh 优先检查 X-Forwarded-For 头部
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// @en X-Forwarded-For can contain multiple IPs separated by comma
		// @zh X-Forwarded-For 可能包含多个用逗号分隔的IP
		for ip := range strings.SplitSeq(xff, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" && !utils.IsPrivateIP(ip) {
				ips = append(ips, ip)
			}
		}
	}

	// @en check X-Real-IP header
	// @zh 检查 X-Real-IP 头部
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		xri = strings.TrimSpace(xri)
		if xri != "" && !utils.IsPrivateIP(xri) {
			ips = append(ips, xri)
		}
	}

	// @en check CF-Connecting-IP header (Cloudflare)
	// @zh 检查 CF-Connecting-IP 头部 (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		cfip = strings.TrimSpace(cfip)
		if cfip != "" && !utils.IsPrivateIP(cfip) {
			ips = append(ips, cfip)
		}
	}

	// @en check X-Forwarded header
	// @zh 检查 X-Forwarded 头部
	if xf := r.Header.Get("X-Forwarded"); xf != "" {
		xf = strings.TrimSpace(xf)
		if xf != "" && !utils.IsPrivateIP(xf) {
			ips = append(ips, xf)
		}
	}

	// @en fallback to RemoteAddr
	// @zh 回退到 RemoteAddr
	if len(ips) == 0 {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// @en if SplitHostPort fails, use RemoteAddr as is
			// @zh 如果 SplitHostPort 失败，直接使用 RemoteAddr
			ips = append(ips, r.RemoteAddr)
		} else {
			ips = append(ips, host)
		}
	}

	return utils.RemoveDuplicates(ips)
}

// @en get uploaded files by key, supports both single and multiple file uploads
//
// @zh 根据键获取上传的文件，支持单文件和批量文件上传
func (r *Request) GetFile(key string, allowType []string) []*multipart.FileHeader {
	// @en check if request is multipart form
	// @zh 检查请求是否为多部分表单
	if r.MultipartForm == nil {
		return nil
	}

	// @en get files from form by key
	// @zh 从表单中根据键获取文件
	files := r.MultipartForm.File[key]
	if len(files) == 0 {
		return nil
	}

	var validFiles []*multipart.FileHeader

	// @en filter files by size limit
	// @zh 根据大小限制过滤文件
	for _, file := range files {
		// @en check if file type is allowed
		// @zh 检查文件类型是否允许
		if len(allowType) > 0 {
			// @en open file to read bytes for type detection
			// @zh 打开文件读取字节用于类型检测
			fileReader, err := file.Open()
			if err != nil {
				continue
			}

			// @en read first 512 bytes for type detection
			// @zh 读取前512字节用于类型检测
			buffer := make([]byte, 512)
			n, err := fileReader.Read(buffer)
			fileReader.Close() // @en close file immediately after reading

			if err != nil && err != io.EOF {
				continue
			}

			// @en detect file type by bytes
			// @zh 通过字节检测文件类型
			contentType := utils.GetFileTypeByBytes(buffer[:n])
			if !utils.Contains(allowType, contentType) {
				continue
			}
		}

		validFiles = append(validFiles, file)
	}

	return validFiles
}

// @en get headers
//
// @zh 获取请求头
func (r *Request) GetHeaders() http.Header {
	return r.Header
}

// @en get header
//
// @zh 获取请求头
func (r *Request) GetHeader(key string) string {
	return r.Header.Get(key)
}

// @en check if websocket is connected
//
// @zh 判断是否是WebSocket连接
func (r *Request) IsWebsocket() bool {
	return r.Method == "GET" &&
		r.GetHeader("Connection") == "Upgrade" &&
		r.GetHeader("Upgrade") == "websocket" &&
		r.GetHeader("Sec-WebSocket-Key") != "" &&
		r.GetHeader("Sec-WebSocket-Version") != ""
}
