package configx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ruomm/goxframework/gox/loggerx"
	"github.com/ruomm/goxiris/xiris/traceutils"
	"reflect"
	"strings"
)

// 遇到数组如何处理 0.放置在外层data_list节点，1.放置在data节点 2.放置在内层data_list节点
type LIST_DATA_MODE int

const (
	LIST_DATA_MODE_OUT  LIST_DATA_MODE = 0
	LIST_DATA_MODE_SELF LIST_DATA_MODE = 1
	LIST_DATA_MODE_IN   LIST_DATA_MODE = 2
)

var listDataMode LIST_DATA_MODE = LIST_DATA_MODE_SELF

// 错误类型统一判断
var typ_error = reflect.TypeOf((*error)(nil)).Elem()
var shortName_ParamError = getType((*CommonParamError)(nil))

func getType(myvar interface{}) string {
	t := reflect.TypeOf(myvar)
	var tmpArr = strings.Split(t.String(), ".")
	shortName := tmpArr[len(tmpArr)-1]
	return strings.ToLower(shortName)
}

// 遇到数组如何处理 0.放置在外层data_list节点，1.放置在data节点 2.放置在内层data_list节点
func ConfigListDataMode(dataMode LIST_DATA_MODE) {
	listDataMode = dataMode
}

type CommonCoreError struct {
	ErrorCode int
	Message   string
}

// 实现error接口的Error方法
func (e *CommonCoreError) Error() string {
	return fmt.Sprintf("Error Code: %d, Message: %s", e.ErrorCode, e.Message)
}

type CommonParamError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
type CommonResult struct {
	//checkOk  bool               // 参数校验是否通过
	//TraceId  string             `json:"traceId,omitempty" newtag:"traceId"`
	Code     int                `json:"code" newtag:"code"`
	Message  string             `json:"message" newtag:"message"`         //用户查看的信息，可读性更强
	Errors   []CommonParamError `json:"errors,omitempty" newtag:"errors"` //打印日志的信息，携带错误详情，便于追查问题
	Data     interface{}        `json:"data,omitempty" newtag:"-"`
	DataList interface{}        `json:"data_list,omitempty" newtag:"-"`
}
type CommonDataWithList struct {
	DataList interface{} `json:"data_list,omitempty" newtag:"-"`
}
type CommonResultType interface {
	string | CommonParamError
}

var CheckResultOk = &CommonResult{Code: ERROR_CODE_OK, Message: "check success"}

func (c CommonResult) Success() bool {
	return c.Code == ERROR_CODE_OK
}

//func (c CommonResult) CheckSuccess() bool {
//	return c.checkOk || c.Code == common.ERROR_CODE_OK
//}

func ToOkResult(pCtx *context.Context, datas ...interface{}) *CommonResult {
	return constructCommonResult(pCtx, ERROR_CODE_OK, datas...)
}

func ToFailResult(pCtx *context.Context, code int, datas ...interface{}) *CommonResult {
	return constructCommonResult(pCtx, code, datas...)
}

func constructCommonResult(pCtx *context.Context, respCode int, datas ...interface{}) *CommonResult {
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
				return nil
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
			if strings.HasSuffix(actualTypeName, shortName_ParamError) {
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
				if listDataMode == LIST_DATA_MODE_OUT {
					datalist = datas[i]
				} else if listDataMode == LIST_DATA_MODE_SELF {
					data = datas[i]
				} else {
					data = &CommonDataWithList{DataList: datas[i]}
				}
			}
		} else if actualKind == reflect.Struct {
			actualTypeName := strings.ToLower(actualValue.String())
			if strings.HasSuffix(actualTypeName, shortName_ParamError) {
				tmpErrorDetail := actualValue.Interface().(CommonParamError)
				errorDetails = append(errorDetails, tmpErrorDetail)
			} else if strings.HasSuffix(actualTypeName, "vo") || strings.HasSuffix(actualTypeName, "data") || strings.HasSuffix(actualTypeName, "resp") || strings.HasSuffix(actualTypeName, "result") {
				data = datas[i]
			}
		}
	}
	if len(message) <= 0 {
		if code == ERROR_CODE_OK {
			message = RESP_MSG_OK
		} else if code == ERROR_CODE_OK_AJAX {
			message = RESP_MSG_OK_AJAX
		} else if code == ERROR_CODE_PARAM_CHECK {
			message = "请求参数错误"
		} else if code == ERROR_CODE_TOKEN_INVALID {
			message = "用户鉴权失败，请重新登陆"
		} else if code == ERROR_CODE_REFUSE {
			message = "操作被拒绝，可能权限不足"
		} else if code == ERROR_CODE_NOT_EXIST {
			message = "查找的资源不存在"
		} else if code == ERROR_CODE_DB_CORE {
			message = "资源处理错误"
		} else if code == ERROR_CODE_FILE_CORE {
			message = "文件读写错误"
		} else if code == ERROR_CODE_THRID_GATEWAY {
			message = "第三方授权失败"
		} else if code == ERROR_CODE_INTERNAL_ERROR {
			message = "服务器内部错误"
		} else if code == ERROR_CODE_UNABLE_HANDLE {
			message = "请求的内容无法处理，请检查请求数据格式"
		} else if code == ERROR_CODE_UNDEFINED {
			message = "未定义错误"
		} else if code == ERROR_CODE_EXCEPTION {
			message = "运行异常"
		} else {
			message = "未定义错误"
		}
	}
	result := &CommonResult{
		Code:     code,
		Message:  message,
		Errors:   errorDetails,
		Data:     data,
		DataList: datalist,
	}
	if code == ERROR_CODE_OK || code == ERROR_CODE_OK_AJAX {
		loggerx.Info(commonResultPrint(pCtx, result))
	} else {
		loggerx.Error(commonResultPrint(pCtx, result))
	}
	// 去除打印时候的traceId
	return result
}

func commonResultPrint(pCtx *context.Context, result *CommonResult) string {
	var build strings.Builder
	if result.Code == ERROR_CODE_OK || result.Code == ERROR_CODE_OK_AJAX {
		logData, _ := json.Marshal(result)
		build.WriteString("请求成功：" + string(logData))
	} else {
		logData, _ := json.Marshal(result)
		build.WriteString("请求失败：" + string(logData))
	}
	timeStatsPrint := traceutils.TraceTimePrint(pCtx)
	if timeStatsPrint != "" {
		build.WriteString(" || ")
		build.WriteString(timeStatsPrint)
	}
	build.WriteString(" || ")
	build.WriteString("traceId：")
	build.WriteString(traceutils.TraceIdFromContext(pCtx))
	return build.String()
}

type CommonPageReqest struct {
	Page     int `json:"page" xreq_param:"page" validate:"min=1" xvalid_error:"当前页数必须为大于0的整数"`
	PageSize int `json:"page" xreq_param:"page_size" validate:"min=1,max=100" xvalid_error:"页记录数必须在[1,100]以内"`
}
type CommonPageVO struct {
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
	LastFlag bool        `json:"last_flag"`
	DataList interface{} `json:"data_list,omitempty"`
}
