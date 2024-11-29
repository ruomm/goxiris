/**
 * @copyright www.ruomm.com
 * @author 牛牛-wanruome@126.com
 * @create 2024/11/27 16:14
 * @version 1.0
 */
package irisutils

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/ruomm/goxframework/gox/corex"
	"io"
	"net/url"
	"os"
	"strings"
)

// iris发送data数据
func SendResponseBytes(irisCtx iris.Context, data []byte, fileName string, inlineMode bool) (int, string, error) {
	if len(fileName) <= 0 {
		return iris.StatusBadRequest, "文件名必须填写", errors.New("文件名必须填写")
	}
	//if len(data) <= 0 {
	//	return iris.StatusBadRequest, "文件内容为空", errors.New("文件内容为空")
	//}
	showFileName := corex.GetFileName(fileName)

	showFileSize := len(data)
	//if showFileSize <= 0 {
	//	fmt.Println("download file is empty")
	//	return iris.StatusNotFound, "文件为空", errors.New("文件为空")
	//}
	respWriter := irisCtx.ResponseWriter()
	if inlineMode {
		//text/html ： HTML格式
		//text/plain ：纯文本格式
		//text/xml ： XML格式
		//image/gif ：gif图片格式
		//image/jpeg ：jpg图片格式
		//image/png：png图片格式
		showFileExt := strings.ToLower(corex.GetFileExtension(showFileName))
		if showFileExt == "jpg" || showFileExt == "jpeg" {
			respWriter.Header().Set("Content-Type", "image/jpeg")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else if showFileExt == "png" {
			respWriter.Header().Set("Content-Type", "image/png")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else if showFileExt == "htm" || showFileName == "html" {
			respWriter.Header().Set("Content-Type", "text/html")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else if showFileExt == "txt" || showFileName == "text" {
			respWriter.Header().Set("Content-Type", "text/plain")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else if showFileExt == "gif" {
			respWriter.Header().Set("Content-Type", "image/gif")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else {
			respWriter.Header().Set("Content-Type", "application/octet-stream")
			respWriter.Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		}
	} else {
		respWriter.Header().Set("Content-Type", "application/octet-stream")
		respWriter.Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
	}
	respWriter.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	respWriter.Header().Set("Content-Length", corex.Int64ToStr(int64(showFileSize)))
	_, errW := respWriter.Write(data)
	if errW != nil {
		return iris.StatusInternalServerError, "IO读写错误", errW
	}
	// 清空缓冲区，进行缓冲区内容落到磁盘
	respWriter.Flush()
	return iris.StatusOK, "", nil
}

// iris发送文件数据
func SendResponseFile(irisCtx iris.Context, filePath string, inlineMode bool) (int, string, error) {
	showFileName := corex.GetFileName(filePath)
	// 判断文件是否存在
	fileInfo, errFileInfo := os.Stat(filePath)
	if errFileInfo != nil {
		fmt.Println("download file not exist", errFileInfo)
		return iris.StatusNotFound, "文件不存在", errFileInfo
	}
	if nil == fileInfo || fileInfo.IsDir() {
		fmt.Println("download file not exist")
		return iris.StatusNotFound, "文件不存在", errors.New("文件不存在")
	}
	showFileSize := fileInfo.Size()
	//if showFileSize <= 0 {
	//	fmt.Println("download file is empty")
	//	return iris.StatusNotFound, "文件为空", errors.New("文件为空")
	//}

	fiR, errOpenR := os.Open(filePath)
	if errOpenR != nil {
		fmt.Println("open file error: ", errOpenR)
		return iris.StatusNotFound, "文件打开失败", errOpenR
	}
	defer fiR.Close()
	respWriter := irisCtx.ResponseWriter()
	if inlineMode {
		//text/html ： HTML格式
		//text/plain ：纯文本格式
		//text/xml ： XML格式
		//image/gif ：gif图片格式
		//image/jpeg ：jpg图片格式
		//image/png：png图片格式
		showFileExt := strings.ToLower(corex.GetFileExtension(showFileName))
		if showFileExt == "jpg" || showFileExt == "jpeg" {
			respWriter.Header().Set("Content-Type", "image/jpeg")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else if showFileExt == "png" {
			respWriter.Header().Set("Content-Type", "image/png")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else if showFileExt == "htm" || showFileName == "html" {
			respWriter.Header().Set("Content-Type", "text/html")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else if showFileExt == "txt" || showFileName == "text" {
			respWriter.Header().Set("Content-Type", "text/plain")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else if showFileExt == "gif" {
			respWriter.Header().Set("Content-Type", "image/gif")
			respWriter.Header().Set("Content-Disposition", "inline; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		} else {
			respWriter.Header().Set("Content-Type", "application/octet-stream")
			respWriter.Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
		}
	} else {
		respWriter.Header().Set("Content-Type", "application/octet-stream")
		respWriter.Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(showFileName)+";filename*="+"utf-8''"+url.QueryEscape(showFileName))
	}
	respWriter.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	respWriter.Header().Set("Content-Length", corex.Int64ToStr(showFileSize))
	bufferSize := 2048
	reader := bufio.NewReader(fiR)
	var bRead = make([]byte, bufferSize)
	//hasRead := false
	for {
		nR, errR := reader.Read(bRead)
		if errR != nil && errR != io.EOF {
			return iris.StatusInternalServerError, "IO读写错误", errR
		}
		if nR == 0 {
			break
		}
		_, errW := respWriter.Write(bRead[0:nR])
		if errW != nil {
			return iris.StatusInternalServerError, "IO读写错误", errW
		}
		if nR < bufferSize {
			break
		}
	}
	// 清空缓冲区，进行缓冲区内容落到磁盘
	respWriter.Flush()
	return iris.StatusOK, "", nil
}
