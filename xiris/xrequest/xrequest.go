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
	"github.com/ruomm/goxframework/gox/refxstandard"
	"github.com/ruomm/goxiris/xiris/xresponse"
	"github.com/ruomm/goxiris/xiris/xvalidator"
)

const xRequest_Parse_Param_COMMON = "xreq_param"

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
			return errors.New("解析参数失败")
		}
	} else if "GET" != ctx.Method() {
		ctx.ReadJSON(req)
		//if err != nil {
		//	return &xrespnse.CommonCoreError{ErrorCode: common.ERROR_CODE_PARAM_CHECK, Message: "解析参数失败"}
		//}
	}
	xrefHander := refxstandard.XrefHander(func(origKey string, key string) (interface{}, error) {
		if ctx.URLParamExists(origKey) {
			paramVal := ctx.URLParam(origKey)
			return paramVal, nil
		} else if ctx.Params().Exists(origKey) {
			paramVal := ctx.Params().GetString(origKey)
			return paramVal, nil
		} else {
			return nil, nil
		}
	})
	errG, transFailsKeys := refxstandard.XRefHandlerCopy(xrefHander, req, refxstandard.XrefOptTag(xRequest_Parse_Param_COMMON), refxstandard.XrefOptCheckUnsigned(true))
	if errG != nil || len(transFailsKeys) > 0 {
		return errors.New("解析参数失败")
	} else {
		return nil
	}

}
