package xjwt

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// 依据contextPath获取接口接口路径
func parseUrlWithContextPath(webContextPath string, uri string) (string, error) {
	if webContextPath == "" {
		if uri == "" {
			return "/", nil
		} else if strings.HasPrefix(uri, "/") {
			return uri, nil
		} else {
			return "/" + uri, nil
		}
	} else if uri == "" {
		if webContextPath == "" {
			return "/", nil
		} else if strings.HasPrefix(webContextPath, "/") {
			return webContextPath, nil
		} else {
			return "/" + webContextPath, nil
		}
	} else if strings.HasPrefix(uri, "/") {
		return uri, nil
	} else {
		resultUrl, err := url.JoinPath(webContextPath, uri)
		if strings.HasPrefix(resultUrl, "/") {
			return resultUrl, err
		} else {
			return "/" + resultUrl, err
		}

	}
}

func getAbsDir(relativePath string) string {
	if len(relativePath) <= 0 {
		return filepath.Dir(os.Args[0])
	} else if xJwtIsAbsDir(relativePath) {
		return relativePath
	} else {
		dir := filepath.Dir(os.Args[0])
		return path.Join(dir, relativePath)
	}
}

func xJwtIsAbsDir(path string) bool {
	if len(path) <= 0 {
		return false
	} else if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return true
	} else {
		indexSpec := strings.Index(path, ":")
		if indexSpec >= 0 && indexSpec < 12 {
			return true
		} else {
			return false
		}
	}
}
