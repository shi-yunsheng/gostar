package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/shi-yunsheng/gostar/date"
	"github.com/shi-yunsheng/gostar/utils"
)

// 下载配置
type Download struct {
	// 最大下载速度，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB
	MaxSpeedLimit string
	// 允许下载的单个文件最大大小，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB
	MaxSize string
}

// 上传配置
type Upload struct {
	// 最大上传速度，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB
	MaxSpeedLimit string
	// 允许上传的单个文件最大大小，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB
	MaxSize string
	// 是否禁用随机名称，禁用后，文件名将保持不变
	DisableRandomName bool
	// 上传文件的键名，默认"file"
	FileKey string
	// 上传后的回调，参数为上传成功的文件路径，可以在此处进行后续处理
	Callback func(w *Response, r Request, filepaths []string)
}

type Static struct {
	// 静态文件路径
	Path string
	// 是否允许目录浏览
	AllowDir bool
	// 允许类型，如：jpg, png, gif, mp4, mp3, etc.
	AllowType []string
	// 下载配置
	Download *Download
	// 上传配置
	Upload *Upload
}

// 带限速下载服务器
func ThrottledDownloadServer(rootDir http.FileSystem, bytesPerSecond int64) Handler {
	return func(w *Response, r Request) any {
		filePath := r.GetParam("__filepath__").(string)
		file, err := rootDir.Open(filePath)
		if err != nil {
			InternalServerError(w, r, err)
			return nil
		}
		defer file.Close()

		reader := utils.NewThrottledReader(file, bytesPerSecond)
		http.ServeContent(w, r.Request, filePath, time.Time{}, reader)
		return nil
	}
}

var (
	// 限速下载服务器
	throttledDownloadServer Handler
	// 文件键名
	fileKey = "file"
	// 上传互斥锁
	uploadMutex sync.Mutex
)

// 静态文件处理器
func StaticServer(handler Handler, staticConfig *Static) Handler {
	return func(w *Response, r Request) any {
		if staticConfig.Path == "" || !utils.IsDir(staticConfig.Path) {
			InternalServerError(w, r, fmt.Errorf("directory does not exist or is not configured"))
			return nil
		}

		if handler != nil {
			handler(w, r)
		}

		if w.GetEarlyBreak() {
			return nil
		}

		filePath := r.GetParam("__filepath__")
		if filePath, ok := filePath.(string); ok && filePath != "" {
			filepath := utils.JoinPath(staticConfig.Path, filePath)
			// 如果允许类型设置，则检查文件类型是否允许
			allowTypeLen := len(staticConfig.AllowType)
			if allowTypeLen > 0 {
				fileType := utils.GetFileType(filepath, true)

				if !utils.Contains(staticConfig.AllowType, fileType) {
					Forbidden(w, r)
					return nil
				}
			}
			// 如果文件不存在，则返回404
			if !(utils.IsExists(filepath) && utils.IsFile(filepath)) {
				NotFound(w, r)
				return nil
			}
			// 如果配置了下载，则下载文件
			if staticConfig.Download != nil {
				allowGet, err := func() (bool, error) {
					// 如果最大大小设置，则检查文件大小是否小于最大大小
					if staticConfig.Download.MaxSize != "" {
						maxSize, err := utils.ParseSize(staticConfig.Download.MaxSize)
						if err != nil {
							return false, err
						}
						if maxSize > 0 && utils.GetFileSize(filepath) > maxSize {
							return false, nil
						}
					}

					return true, nil
				}()

				if err != nil {
					InternalServerError(w, r, err)
					return nil
				}

				if !allowGet {
					Forbidden(w, r)
					return nil
				}

				filename := utils.GetFileName(filepath)
				w.Header().Set("Content-Disposition", "attachment; filename="+filename)
				w.Header().Set("Content-Type", utils.GetFileMime(filepath)+"; charset=utf-8")
				w.Header().Set("Content-Length", strconv.FormatInt(utils.GetFileSize(filepath), 10))
				// 如果最大速度限制设置，则使用限速下载服务器
				if staticConfig.Download.MaxSpeedLimit != "" {
					speedLimit, err := utils.ParseSize(staticConfig.Download.MaxSpeedLimit)
					if err != nil {
						InternalServerError(w, r, err)
						return nil
					}
					if speedLimit > 0 {
						if throttledDownloadServer == nil {
							throttledDownloadServer = ThrottledDownloadServer(http.Dir(staticConfig.Path), speedLimit)
						}
						throttledDownloadServer(w, r)
						return nil
					}
				}
			}
			http.ServeFile(w, r.Request, filepath)
			return nil
		}
		// 如果配置了上传，并且有携带文件，则上传文件
		if staticConfig.Upload != nil {
			if staticConfig.Upload.FileKey != "" {
				fileKey = staticConfig.Upload.FileKey
			}

			files := r.GetFile(fileKey, staticConfig.AllowType)

			if len(files) > 0 {
				filepaths := make([]string, len(files))

				maxSize, err := utils.ParseSize(staticConfig.Upload.MaxSize)
				if err != nil {
					InternalServerError(w, r, err)
					return nil
				}
				// 创建日期目录
				dateDir := utils.JoinPath(staticConfig.Path, date.GetToday())
				uploadMutex.Lock()
				utils.Mkdir(dateDir)
				uploadMutex.Unlock()

				for i, file := range files {
					// 如果指定了最大大小，检查文件大小
					if maxSize > 0 && file.Size > maxSize {
						InternalServerError(w, r, fmt.Errorf("file size exceeds max size"))
						return nil
					}
					// 获取文件名
					if !staticConfig.Upload.DisableRandomName {
						file.Filename = utils.NewUUID()
					}

					filepath := utils.JoinPath(dateDir, file.Filename)
					// 如果文件存在，则添加时间戳和随机字符串，防止文件名爆破
					uploadMutex.Lock()
					if utils.IsExists(filepath) {
						file.Filename = file.Filename + "_" + strconv.FormatInt(date.GetTimestamp(true), 10) + utils.GetRandomString(8)
					}
					uploadMutex.Unlock()
					// 如果最大速率限制设置，则限制上传速率
					if staticConfig.Upload.MaxSpeedLimit != "" {
						speedLimit, err := utils.ParseSize(staticConfig.Upload.MaxSpeedLimit)
						if err != nil {
							InternalServerError(w, r, err)
							return nil
						}
						if speedLimit > 0 {
							err = utils.SaveFile(file, filepath, speedLimit)
							if err != nil {
								InternalServerError(w, r, err)
								return nil
							}
						}
					} else {
						err = utils.SaveFile(file, filepath, 0)
						if err != nil {
							InternalServerError(w, r, err)
							return nil
						}
					}
					// 存储上传的文件路径
					filepaths[i] = filepath
				}

				if staticConfig.Upload.Callback != nil {
					staticConfig.Upload.Callback(w, r, filepaths)
				}

				return nil
			}
		}
		// 如果允许目录浏览，则浏览目录
		if staticConfig.AllowDir {
			http.ServeFile(w, r.Request, staticConfig.Path)
			return nil
		}

		Forbidden(w, r)
		return nil
	}
}
