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
	"github.com/ruomm/goxframework/gox/refx"
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

type XRequestHander func(ctx iris.Context, key string) (string, error)

var xRequestHandler refx.XrefHandler = nil

// 配置XRequestHander
func ConfigRequestHandler(handler refx.XrefHandler) {
	xRequestHandler = handler
}

func XRequestParse(pCtx iris.Context, req interface{}) (error, *[]xresponse.ParamError) {
	// 解析参数
	err := xReq_parse(pCtx, req)
	if err != nil {
		return err, nil
	}
	// 验证参数
	return xvalidator.XValidator(req)
}
func xReq_parse(ctx iris.Context, req interface{}) error {
	//if "POST" == ctx.Method() || "PUT" == ctx.Method() {
	//	err := ctx.ReadJSON(req)
	//	if err != nil {
	//		return errors.New("解析JSON参数失败")
	//	}
	//} else if "GET" != ctx.Method() {
	//	ctx.ReadJSON(req)
	//}

	if "GET" != ctx.Method() {
		body, err := ctx.GetBody()
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
		if ctx.Params().Exists(origKey) {
			paramVal := ctx.Params().GetString(origKey)
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
		if ctx.URLParamExists(origKey) {
			paramVal := ctx.URLParam(origKey)
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
			paramVal := ctx.ResponseWriter().Header().Get(origKey)
			if len(paramVal) > 0 {
				return paramVal, nil
			} else {
				return nil, nil
			}
		} else {
			paramVal := ctx.GetHeader(origKey)
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
		// 解析refx参数
		errGRefx, transFailsKeysRefx := refxstandard.XRefHandlerCopy(xrefHanderQuery, req, refxstandard.XrefOptTag(xRequest_Parse_refx), refxstandard.XrefOptCheckUnsigned(true))
		if errGRefx != nil || len(transFailsKeysRefx) > 0 {
			return errors.New("解析refx参数失败")
		}
	}
	return nil
}

func xTagContainKey(tagValue string, key string) bool {
	return corex.TagOptions(tagValue).Contains(key)
}
