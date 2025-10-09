package utils

import (
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
)

// @en check if file or directory exists
//
// @zh 检查文件或目录是否存在
func IsExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// @en check if path is a directory
//
// @zh 检查路径是否为目录
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// @en check if path is a file
//
// @zh 检查路径是否为文件
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// @en join path
//
// @zh 拼接路径
func JoinPath(path ...string) string {
	return filepath.Join(path...)
}

// @en get file name from path
//
// @zh 从路径中获取文件名
func GetFileName(path string) string {
	return filepath.Base(path)
}

// @en Get the file type using the standard library DetectContentType.
// If useExtension is set, the file extension will be used first.
//
// @zh 使用标准库DetectContentType获取文件类型。
// 如果设置了使用扩展名，则优先使用扩展名。
func GetFileType(path string, useExtension ...bool) string {
	mime, err := mimetype.DetectFile(path)
	if len(useExtension) > 0 && useExtension[0] || err != nil {
		ext := filepath.Ext(path)
		if ext != "" {
			return ext[1:]
		}
		return "unknown"
	}
	return mime.Extension()[1:]
}

// @en get file type by bytes
//
// @zh 通过字节获取文件类型
func GetFileTypeByBytes(bytes []byte) string {
	mime := mimetype.Detect(bytes)
	return mime.Extension()[1:]
}

// @en get file mime type
//
// @zh 获取文件MIME类型
func GetFileMime(path string) string {
	mime, err := mimetype.DetectFile(path)
	if err != nil {
		return "application/octet-stream"
	}
	return mime.String()
}

// @en get file mime type by bytes
//
// @zh 通过字节获取文件MIME类型
func GetFileMimeByBytes(bytes []byte) string {
	mime := mimetype.Detect(bytes)
	return mime.String()
}

// @en get file size
//
// @zh 获取文件大小
func GetFileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

// @en create directory
//
// @zh 创建目录
func Mkdir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// @en save file
//
// @zh 保存文件
func SaveFile(file *multipart.FileHeader, filepath string, bytesPerSecond int64) error {
	// @en open source file
	// @zh 打开源文件
	src, err := file.Open()
	if err != nil {
		return err
	}

	// @en create destination file
	// @zh 创建目标文件
	dst, err := os.Create(filepath)
	if err != nil {
		src.Close()
		return err
	}

	defer func() {
		src.Close()
		dst.Close()
	}()

	// @en copy file content with throttled reader
	// @zh 复制文件内容并使用限速读取器
	if bytesPerSecond > 0 {
		reader := NewThrottledReader(src, bytesPerSecond)
		_, err = dst.ReadFrom(reader)
	} else {
		_, err = dst.ReadFrom(src)
	}

	return err
}
