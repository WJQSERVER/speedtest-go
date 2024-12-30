package web

import (
	"net/http"
	"os"
	"path/filepath"
)

// Credit: https://stackoverflow.com/questions/49589685/good-way-to-disable-directory-listing-with-http-fileserver-in-go
/*
type justFilesFilesystem struct {
	fs http.FileSystem
	// readDirBatchSize - configuration parameter for `Readdir` func
	readDirBatchSize int
}

func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return neuteredStatFile{File: f, readDirBatchSize: fs.readDirBatchSize}, nil
}

type neuteredStatFile struct {
	http.File
	readDirBatchSize int
}

func (e neuteredStatFile) Stat() (os.FileInfo, error) {
	s, err := e.File.Stat()
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
	LOOP:
		for {
			fl, err := e.File.Readdir(e.readDirBatchSize)
			switch err {
			case io.EOF:
				break LOOP
			case nil:
				for _, f := range fl {
					if f.Name() == "index.html" {
						return s, err
					}
				}
			default:
				return nil, err
			}
		}
		return nil, os.ErrNotExist
	}
	return s, err
}
*/

type justFilesFilesystem struct {
	fs http.FileSystem
}

func (jfs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := jfs.fs.Open(name)
	if err != nil {
		return nil, err
	}

	// 检查是否为目录
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		// 构造目录下index.html的路径
		indexFilePath := filepath.Join(name, "index.html")
		indexFile, err := jfs.fs.Open(indexFilePath)
		if err != nil {
			// 如果打开index.html失败，则返回目录不存在的错误
			return nil, os.ErrNotExist
		}

		// 返回index.html文件
		return indexFile, nil
	}

	// 返回文件
	return f, nil
}
