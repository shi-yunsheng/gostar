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

// @en download configuration
//
// @zh 下载配置
type Download struct {
	// @en maximum download speed, supports units: B, KB, MB, GB, TB, e.g.: 1KB, 2.5MB, 1GB
	//
	// @zh 最大下载速度，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB
	MaxSpeedLimit string
	// @en maximum file size for single download, supports units: B, KB, MB, GB, TB, e.g.: 1KB, 2.5MB, 1GB
	//
	// @zh 允许下载的单个文件最大大小，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB
	MaxSize string
}

// @en upload configuration
//
// @zh 上传配置
type Upload struct {
	// @en maximum upload speed, supports units: B, KB, MB, GB, TB, e.g.: 1KB, 2.5MB, 1GB
	//
	// @zh 最大上传速度，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB
	MaxSpeedLimit string
	// @en maximum file size for single upload, supports units: B, KB, MB, GB, TB, e.g.: 1KB, 2.5MB, 1GB
	//
	// @zh 允许上传的单个文件最大大小，支持单位：B, KB, MB, GB, TB，如：1KB, 2.5MB, 1GB
	MaxSize string
	// @en disable random file name generation
	//
	// @zh 是否禁用随机名称，禁用后，文件名将保持不变
	DisableRandomName bool
	// @en form field key for file upload, default "file"
	//
	// @zh 上传文件的键名，默认"file"
	FileKey string
	// @en upload success callback, parameter is the uploaded file path, can be used to process the subsequent process
	//
	// @zh 上传后的回调，参数为上传成功的文件路径，可以在此处进行后续处理
	Callback func(w *Response, r Request, filepaths []string)
}

type Static struct {
	// @en static file path
	//
	// @zh 静态文件路径
	Path string
	// @en allow directory browsing
	//
	// @zh 是否允许目录浏览
	AllowDir bool
	// @en allowed file types, e.g.: jpg, png, gif, mp4, mp3, etc.
	//
	// @zh 允许类型，如：jpg, png, gif, mp4, mp3, etc.
	AllowType []string
	// @en download configuration
	//
	// @zh 下载配置
	Download *Download
	// @en upload configuration
	//
	// @zh 上传配置
	Upload *Upload
}

// @en download file handler with throttled reader
//
// @zh 带限速下载服务器
func ThrottledDownloadServer(rootDir http.FileSystem, bytesPerSecond int64) Handler {
	return func(w *Response, r Request) {
		filePath := r.GetParam("__filepath__").(string)
		file, err := rootDir.Open(filePath)
		if err != nil {
			InternalServerError(w, r, err)
			return
		}
		defer file.Close()

		reader := utils.NewThrottledReader(file, bytesPerSecond)
		http.ServeContent(w, r.Request, filePath, time.Time{}, reader)
	}
}

var (
	// @en throttled download server
	//
	// @zh 限速下载服务器
	throttledDownloadServer Handler
	// @en file key
	//
	// @zh 文件键名
	fileKey = "file"
	// @en upload mutex
	//
	// @zh 上传互斥锁
	uploadMutex sync.Mutex
)

// @en static file handler
//
// @zh 静态文件处理器
func StaticServer(handler Handler, staticConfig *Static) Handler {
	return func(w *Response, r Request) {
		if staticConfig.Path == "" || !utils.IsDir(staticConfig.Path) {
			InternalServerError(w, r, fmt.Errorf("directory does not exist or is not configured"))
			return
		}

		if handler != nil {
			handler(w, r)
		}

		if w.GetEarlyBreak() {
			return
		}

		filePath := r.GetParam("__filepath__")
		if filePath, ok := filePath.(string); ok && filePath != "" {
			filepath := utils.JoinPath(staticConfig.Path, filePath)

			// @en if allow type is set, check if file type is allowed
			// @zh 如果允许类型设置，则检查文件类型是否允许
			allowTypeLen := len(staticConfig.AllowType)
			if allowTypeLen > 0 {
				fileType := utils.GetFileType(filepath, true)

				if !utils.Contains(staticConfig.AllowType, fileType) {
					Forbidden(w, r)
					return
				}
			}

			// @en if file does not exist, return 404
			// @zh 如果文件不存在，则返回404
			if !(utils.IsExists(filepath) && utils.IsFile(filepath)) {
				NotFound(w, r)
				return
			}

			// @en if download config is set, download file
			// @zh 如果配置了下载，则下载文件
			if staticConfig.Download != nil {
				allowGet, err := func() (bool, error) {
					// @en if max size is set, check if file size is less than max size
					// @zh 如果最大大小设置，则检查文件大小是否小于最大大小
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
					return
				}

				if !allowGet {
					Forbidden(w, r)
					return
				}

				filename := utils.GetFileName(filepath)
				w.Header().Set("Content-Disposition", "attachment; filename="+filename)
				w.Header().Set("Content-Type", utils.GetFileMime(filepath)+"; charset=utf-8")
				w.Header().Set("Content-Length", strconv.FormatInt(utils.GetFileSize(filepath), 10))

				// @en if max speed limit is set, use throttled download server
				// @zh 如果最大速度限制设置，则使用限速下载服务器
				if staticConfig.Download.MaxSpeedLimit != "" {
					speedLimit, err := utils.ParseSize(staticConfig.Download.MaxSpeedLimit)
					if err != nil {
						InternalServerError(w, r, err)
						return
					}
					if speedLimit > 0 {
						if throttledDownloadServer == nil {
							throttledDownloadServer = ThrottledDownloadServer(http.Dir(staticConfig.Path), speedLimit)
						}
						throttledDownloadServer(w, r)
						return
					}
				}
			}
			http.ServeFile(w, r.Request, filepath)
			return
		}

		// @en if upload config is set, upload file
		// @zh 如果配置了上传，并且有携带文件，则上传文件
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
					return
				}

				// @en create date directory
				// @zh 创建日期目录
				dateDir := utils.JoinPath(staticConfig.Path, date.GetToday())
				uploadMutex.Lock()
				utils.Mkdir(dateDir)
				uploadMutex.Unlock()

				for i, file := range files {
					// @en if maxSize is specified, check file size
					// @zh 如果指定了最大大小，检查文件大小
					if maxSize > 0 && file.Size > maxSize {
						InternalServerError(w, r, fmt.Errorf("file size exceeds max size"))
						return
					}

					// @en get file name
					// @zh 获取文件名
					if !staticConfig.Upload.DisableRandomName {
						file.Filename = utils.NewUUID()
					}

					filepath := utils.JoinPath(dateDir, file.Filename)

					// @en if file exists, add timestamp and random string to prevent file name brute force
					// @zh 如果文件存在，则添加时间戳和随机字符串，防止文件名爆破
					uploadMutex.Lock()
					if utils.IsExists(filepath) {
						file.Filename = file.Filename + "_" + strconv.FormatInt(date.GetTimestamp(true), 10) + utils.GetRandomString(8)
					}
					uploadMutex.Unlock()

					// @en if max speed limit is set, limit upload speed
					// @zh 如果最大速率限制设置，则限制上传速率
					if staticConfig.Upload.MaxSpeedLimit != "" {
						speedLimit, err := utils.ParseSize(staticConfig.Upload.MaxSpeedLimit)
						if err != nil {
							InternalServerError(w, r, err)
							return
						}
						if speedLimit > 0 {
							err = utils.SaveFile(file, filepath, speedLimit)
							if err != nil {
								InternalServerError(w, r, err)
								return
							}
						}
					} else {
						err = utils.SaveFile(file, filepath, 0)
						if err != nil {
							InternalServerError(w, r, err)
							return
						}
					}

					// @en store uploaded file path
					// @zh 存储上传的文件路径
					filepaths[i] = filepath
				}

				if staticConfig.Upload.Callback != nil {
					staticConfig.Upload.Callback(w, r, filepaths)
				}

				return
			}
		}

		// @en if allow directory browsing, browse directory
		// @zh 如果允许目录浏览，则浏览目录
		if staticConfig.AllowDir {
			http.ServeFile(w, r.Request, staticConfig.Path)
			return
		}

		Forbidden(w, r)
	}
}
