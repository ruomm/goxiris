/**
 * @copyright 像衍科技-idr.ai
 * @author 牛牛-研发部-www.ruomm.com
 * @create 2024/1/23 10:01
 * @version 1.0
 */
package configx

import (
	"reflect"
	"strings"
)

type CommonParamError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
type CommonResponse struct {
	MessageMap          map[int]string
	OkCode              int
	OkMessage           string
	shortNameParamError string
}

func GenCommonResponse(okCode int, okMessage string, messageMap map[int]string) *CommonResponse {
	commonResponse := CommonResponse{
		OkCode:     okCode,
		OkMessage:  okMessage,
		MessageMap: messageMap,
	}
	commonResponse.shortNameParamError = getType((*CommonParamError)(nil))
	return &commonResponse
}
func getType(myvar interface{}) string {
	t := reflect.TypeOf(myvar)
	var tmpArr = strings.Split(t.String(), ".")
	shortName := tmpArr[len(tmpArr)-1]
	return strings.ToLower(shortName)
}

func (t *CommonResponse) constructCommonResult2(respCode int, datas ...interface{}) (int, string, []CommonParamError, interface{}, interface{}) {
	var code = respCode
	var message string
	var errorDetails []CommonParamError = nil
	var data interface{} = nil
	var datalist interface{} = nil

	for i := 0; i < len(datas); i++ {
		origVal := datas[i]
		if nil == origVal {
			continue
		}
		// 获取真实的数值
		actualValue := reflect.ValueOf(origVal)
		if actualValue.Kind() == reflect.Pointer || actualValue.Kind() == reflect.Interface {
			if actualValue.IsNil() {
				continue
			}
			actualValue = actualValue.Elem()
		}
		actualKind := actualValue.Kind()
		if actualKind == reflect.Int {
			intType := reflect.TypeOf(int(0))
			if intType != actualValue.Type() {
				actualValue = actualValue.Convert(intType)
			}
			code = actualValue.Interface().(int)
		} else if actualKind == reflect.String {
			stringType := reflect.TypeOf("")
			if stringType != actualValue.Type() {
				actualValue = actualValue.Convert(stringType)
			}
			message = actualValue.Interface().(string)
		} else if actualKind == reflect.Slice {
			actualTypeName := strings.ToLower(actualValue.String())
			if strings.HasSuffix(actualTypeName, t.shortNameParamError) {
				tmpErrorDetails := actualValue.Interface().([]CommonParamError)
				if len(tmpErrorDetails) <= 0 {
					continue
				}
				if errorDetails == nil {
					errorDetails = tmpErrorDetails
				} else {
					for _, tmpErrorDetail := range tmpErrorDetails {
						errorDetails = append(errorDetails, tmpErrorDetail)
					}
				}
			} else if strings.HasSuffix(actualTypeName, "vo") || strings.HasSuffix(actualTypeName, "data") || strings.HasSuffix(actualTypeName, "resp") || strings.HasSuffix(actualTypeName, "result") {
				datalist = datas[i]
			}
		} else if actualKind == reflect.Struct {
			actualTypeName := strings.ToLower(actualValue.String())
			if strings.HasSuffix(actualTypeName, t.shortNameParamError) {
				tmpErrorDetail := actualValue.Interface().(CommonParamError)
				errorDetails = append(errorDetails, tmpErrorDetail)
			} else if strings.HasSuffix(actualTypeName, "vo") || strings.HasSuffix(actualTypeName, "data") || strings.HasSuffix(actualTypeName, "resp") || strings.HasSuffix(actualTypeName, "result") {
				data = datas[i]
			}
		}
	}
	if len(message) <= 0 {
		if code == t.OkCode {
			message = t.OkMessage
		}
		tmpMessage := ""
		if t.MessageMap != nil && len(t.MessageMap) > 0 {
			tmpMessage = t.MessageMap[code]
		}
		if len(tmpMessage) > 0 {
			message = tmpMessage
		} else {
			message = "未定义错误"
		}
	}
	return code, message, errorDetails, data, datalist
}
