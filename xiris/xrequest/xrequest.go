/**
 * @copyright 像衍科技-idr.ai
 * @author 牛牛-研发部-www.ruomm.com
 * @create 2024/1/22 14:48
 * @version 1.0
 */
package xrequest

import (
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
	xRequest_Option_Response = "resp"
)

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
	if "POST" == ctx.Method() || "PUT" == ctx.Method() {
		err := ctx.ReadJSON(req)
		if err != nil {
			return errors.New("解析JSON参数失败")
		}
	} else if "GET" != ctx.Method() {
		ctx.ReadJSON(req)
		//if err != nil {
		//	return &xrespnse.CommonCoreError{ErrorCode: common.ERROR_CODE_PARAM_CHECK, Message: "解析参数失败"}
		//}
	}
	// 解析URI参数
	xrefHanderParam := refxstandard.XrefHander(func(origKey string, key string, cpOpt string) (interface{}, error) {
		if ctx.Params().Exists(origKey) {
			paramVal := ctx.Params().GetString(origKey)
			return paramVal, nil
		} else {
			return nil, nil
		}
	})
	errGParam, transFailsKeysParam := refxstandard.XRefHandlerCopy(xrefHanderParam, req, refxstandard.XrefOptTag(xRequest_Parse_Param), refxstandard.XrefOptCheckUnsigned(true))
	if errGParam != nil || len(transFailsKeysParam) > 0 {
		return errors.New("解析URI参数失败")
	}
	// 解析query参数
	xrefHanderQuery := refxstandard.XrefHander(func(origKey string, key string, cpOpt string) (interface{}, error) {
		if ctx.URLParamExists(origKey) {
			paramVal := ctx.URLParam(origKey)
			return paramVal, nil
		} else {
			return nil, nil
		}
	})
	errGQuery, transFailsKeysQuery := refxstandard.XRefHandlerCopy(xrefHanderQuery, req, refxstandard.XrefOptTag(xRequest_Parse_Query), refxstandard.XrefOptCheckUnsigned(true))
	if errGQuery != nil || len(transFailsKeysQuery) > 0 {
		return errors.New("解析Query参数失败")
	}
	// 解析header参数
	xrefHanderHeader := refxstandard.XrefHander(func(origKey string, key string, cpOpt string) (interface{}, error) {
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
	return nil
}

func xTagContainKey(tagValue string, key string) bool {
	return corex.TagOptions(tagValue).Contains(key)
}
