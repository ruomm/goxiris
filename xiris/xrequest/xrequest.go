/**
 * @copyright 像衍科技-idr.ai
 * @author 牛牛-研发部-www.ruomm.com
 * @create 2024/1/22 14:48
 * @version 1.0
 */
package xrequest

import (
	"encoding/json"
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/ruomm/goxframework/gox/corex"
	"github.com/ruomm/goxframework/gox/refxstandard"
	"github.com/ruomm/goxiris/xiris/xresponse"
	"github.com/ruomm/goxiris/xiris/xvalidator"
)

const (
	//xRequest_Parse_Param_COMMON = "xreq_param"
	xRequest_Parse_Param     = "xreq_param"
	xRequest_Parse_Query     = "xreq_query"
	xRequest_Parse_Header    = "xreq_header"
	xRequest_Parse_refx      = "xreq_refx"
	xRequest_Option_Response = "resp"
)

type XRequestHander func(irisCtx iris.Context, origKey string, key string, cpOpt string) (interface{}, error)

var xRequestHandler XRequestHander = nil

var showFirstError bool = false

// 配置XRequestHander
func ConfigRequestHandler(handler XRequestHander) {
	xRequestHandler = handler
}

// 配置是否显示第一条错误信息
func ConfigshowFirstError(show bool) {
	showFirstError = show
}

func XRequestParse(irisCtx iris.Context, req interface{}) (error, *[]xresponse.ParamError) {
	// 解析参数
	err := xReq_parse(irisCtx, req)
	if err != nil {
		return err, nil
	}
	// 验证参数
	return xvalidator.XValidator(req, showFirstError)
}
func xReq_parse(irisCtx iris.Context, req interface{}) error {
	//if "POST" == irisCtx.Method() || "PUT" == irisCtx.Method() {
	//	err := irisCtx.ReadJSON(req)
	//	if err != nil {
	//		return errors.New("解析JSON参数失败")
	//	}
	//} else if "GET" != irisCtx.Method() {
	//	irisCtx.ReadJSON(req)
	//}

	if "GET" != irisCtx.Method() {
		body, err := irisCtx.GetBody()
		if err != nil {
			return errors.New("读取请求body错误")
		}
		if body != nil && len(body) > 0 {
			errJSON := json.Unmarshal(body, req)
			if errJSON != nil {
				return errors.New("解析JSON参数失败")
			}
		}
	}
	// 解析URI参数
	xrefHanderParam := refxstandard.XrefHandler(func(origKey string, key string, cpOpt string) (interface{}, error) {
		if irisCtx.Params().Exists(origKey) {
			paramVal := irisCtx.Params().GetString(origKey)
			if len(paramVal) > 0 {
				return paramVal, nil
			} else {
				return nil, nil
			}
		} else {
			return nil, nil
		}
	})
	errGParam, transFailsKeysParam := refxstandard.XRefHandlerCopy(xrefHanderParam, req, refxstandard.XrefOptTag(xRequest_Parse_Param), refxstandard.XrefOptCheckUnsigned(true))
	if errGParam != nil || len(transFailsKeysParam) > 0 {
		return errors.New("解析URI参数失败")
	}
	// 解析query参数
	xrefHanderQuery := refxstandard.XrefHandler(func(origKey string, key string, cpOpt string) (interface{}, error) {
		if irisCtx.URLParamExists(origKey) {
			paramVal := irisCtx.URLParam(origKey)
			if len(paramVal) > 0 {
				return paramVal, nil
			} else {
				return nil, nil
			}
		} else {
			return nil, nil
		}
	})
	errGQuery, transFailsKeysQuery := refxstandard.XRefHandlerCopy(xrefHanderQuery, req, refxstandard.XrefOptTag(xRequest_Parse_Query), refxstandard.XrefOptCheckUnsigned(true))
	if errGQuery != nil || len(transFailsKeysQuery) > 0 {
		return errors.New("解析Query参数失败")
	}
	// 解析header参数
	xrefHanderHeader := refxstandard.XrefHandler(func(origKey string, key string, cpOpt string) (interface{}, error) {
		if xTagContainKey(cpOpt, xRequest_Option_Response) {
			paramVal := irisCtx.ResponseWriter().Header().Get(origKey)
			if len(paramVal) > 0 {
				return paramVal, nil
			} else {
				return nil, nil
			}
		} else {
			paramVal := irisCtx.GetHeader(origKey)
			if len(paramVal) > 0 {
				return paramVal, nil
			} else {
				return nil, nil
			}
		}
	})
	errGHeader, transFailsKeysHeader := refxstandard.XRefHandlerCopy(xrefHanderHeader, req, refxstandard.XrefOptTag(xRequest_Parse_Header), refxstandard.XrefOptCheckUnsigned(true))
	if errGHeader != nil || len(transFailsKeysHeader) > 0 {
		return errors.New("解析Header参数失败")
	}
	if nil != xRequestHandler {
		xrefHanderRefx := refxstandard.XrefHandler(func(origKey string, key string, cpOpt string) (interface{}, error) {
			return xRequestHandler(irisCtx, origKey, key, cpOpt)
		})
		// 解析自定义refx参数
		errGRefx, transFailsKeysRefx := refxstandard.XRefHandlerCopy(xrefHanderRefx, req, refxstandard.XrefOptTag(xRequest_Parse_refx), refxstandard.XrefOptCheckUnsigned(true))
		if errGRefx != nil || len(transFailsKeysRefx) > 0 {
			return errors.New("解析自定义refx参数失败")
		}
	}
	return nil
}

func xTagContainKey(tagValue string, key string) bool {
	return corex.TagOptions(tagValue).Contains(key)
}
