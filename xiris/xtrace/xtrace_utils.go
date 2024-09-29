package xtrace

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"strconv"
	"strings"
	"time"
)

type XTraceClient struct {
	// trace信息存储key
	KeyTraceSave string
	// traceId在header中的KEY
	KeyTraceId string
	// ts在header中的KEY
	KeyTs string
	// uri在header中的KEY
	KeyUri string
	// userId在header中的KEY
	KeyUserId string
	// userName在header中的KEY
	KeyUserName string
	// roleId在header中的KEY
	KeyRoleId string
	// 是否从responseHeader里面读取
	ByResponse bool
}

func GenXTraceClient(client XTraceClient) XTraceClient {
	return genXTraceClientCommon(client, false)
}

func GenXTraceClientWithDefault(client XTraceClient) XTraceClient {
	return genXTraceClientCommon(client, true)
}

/*
* keyTrace：trace信息存储key
keyTraceId：traceId在header中的KEY
keyTs：ts在header中的KEY
keyUri：uri在header中的KEY
keyUserId：userId在header中的KEY
keyUserName:userName在header中的KEY
keyRoleId:roleId在header中的KEY
byResponse：是否从responseHeader里面读取
*/
func genXTraceClientCommon(client XTraceClient, defaultKeyMode bool) XTraceClient {
	realKeyTraceSave := client.KeyTraceSave
	if len(realKeyTraceSave) <= 0 {
		if defaultKeyMode {
			realKeyTraceSave = "KEY_TRACE"
		} else {
			realKeyTraceSave = ""
		}
	}
	realKeyTraceId := client.KeyTraceId
	if len(realKeyTraceId) <= 0 {
		if defaultKeyMode {
			realKeyTraceId = "X-IDR-TraceId"
		} else {
			realKeyTraceId = ""
		}
	}
	realKeyTs := client.KeyTs
	if len(realKeyTs) <= 0 {
		if defaultKeyMode {
			realKeyTs = "X-IDR-Ts"
		} else {
			realKeyTs = ""
		}
	}
	realKeyUri := client.KeyUri
	if len(realKeyUri) <= 0 {
		if defaultKeyMode {
			realKeyUri = "X-IDR-Uri"
		} else {
			realKeyUri = ""
		}
	}
	realKeyUserId := client.KeyUserId
	if len(realKeyUserId) <= 0 {
		if defaultKeyMode {
			realKeyUserId = "__user_id"
		} else {
			realKeyUserId = ""
		}
	}
	realKeyUserName := client.KeyUserName
	if len(realKeyUserName) <= 0 {
		if defaultKeyMode {
			realKeyUserName = "__user_name"
		} else {
			realKeyUserName = ""
		}
	}
	realKeyRoleId := client.KeyRoleId
	if len(realKeyRoleId) <= 0 {
		if defaultKeyMode {
			realKeyRoleId = "__user_role"
		} else {
			realKeyRoleId = ""
		}
	}
	xTraceClient := XTraceClient{
		KeyTraceSave: realKeyTraceSave,
		KeyTraceId:   realKeyTraceId,
		KeyTs:        realKeyTs,
		KeyUri:       realKeyUri,
		KeyUserId:    realKeyUserId,
		KeyUserName:  realKeyUserName,
		KeyRoleId:    realKeyRoleId,
		ByResponse:   client.ByResponse,
	}
	return xTraceClient
}

type XTraceInfo struct {
	TraceId  string
	TsArr    []int64
	Events   []string
	Uri      string
	UserId   uint
	UserName string
	RoleId   uint32
}

/*
* 获取一个带TraceInfo的Context环境变量
 */
func (t *XTraceClient) ToTraceContext(irisCtx iris.Context) context.Context {
	traceId := ""
	uriStr := ""
	tsHeader := ""
	userIdStr := ""
	userNameStr := ""
	userRoleIdStr := ""
	if t.ByResponse {
		traceId = irisCtx.ResponseWriter().Header().Get(t.KeyTraceId)
		uriStr = irisCtx.ResponseWriter().Header().Get(t.KeyUri)
		tsHeader = irisCtx.ResponseWriter().Header().Get(t.KeyTs)
		userIdStr = irisCtx.ResponseWriter().Header().Get(t.KeyUserId)
		userNameStr = irisCtx.ResponseWriter().Header().Get(t.KeyUserName)
		userRoleIdStr = irisCtx.ResponseWriter().Header().Get(t.KeyRoleId)
	} else {
		traceId = irisCtx.GetHeader(t.KeyTraceId)
		uriStr = irisCtx.GetHeader(t.KeyUri)
		tsHeader = irisCtx.GetHeader(t.KeyTs)
		userIdStr = irisCtx.GetHeader(t.KeyUserId)
		userNameStr = irisCtx.GetHeader(t.KeyUserName)
		userRoleIdStr = irisCtx.GetHeader(t.KeyRoleId)
	}
	var userId uint = 0
	if len(userIdStr) > 0 {
		tmpUserId, err := strconv.ParseUint(userIdStr, 10, 64)
		if err == nil {
			userId = uint(tmpUserId)
		}
	}
	var roldId uint32 = 0
	if len(userRoleIdStr) > 0 {
		roleId, err := strconv.ParseUint(userRoleIdStr, 10, 64)
		if err == nil {
			roldId = uint32(roleId)
		}
	}
	var tsArr []int64 = nil
	var events []string = nil
	if tsHeader != "" {
		ts, err := strconv.ParseInt(tsHeader, 10, 64)
		if err != nil {
			// 可能字符串 s 不是合法的整数格式，处理错误
			fmt.Println(err)
			tsArr = nil
			events = nil

		} else {
			tsArr = []int64{ts}
			events = nil
		}
	}
	xTraceInfo := XTraceInfo{
		TraceId:  traceId,
		TsArr:    tsArr,
		Events:   events,
		Uri:      uriStr,
		UserId:   userId,
		UserName: userNameStr,
		RoleId:   roldId,
	}
	ctx := context.WithValue(context.Background(), t.KeyTraceSave, &xTraceInfo)
	return ctx
}

/**
* 从Context环境变量中获取traceInfo
 */
func (t *XTraceClient) GetTraceInfo(ctx context.Context) *XTraceInfo {
	if ctx == nil {
		return nil
	}
	pTraceInfoAny := ctx.Value(t.KeyTraceSave)
	if pTraceInfoAny == nil {
		return nil
	}
	pTraceInfo := pTraceInfoAny.(*XTraceInfo)
	return pTraceInfo
}

/**
* 从Context环境变量中清理traceInfo
 */
func (t *XTraceClient) CleanTraceInfoAll(irisCtx iris.Context) {
	t.cleanTraceInfoCommon(irisCtx, true)
}

/**
* 从Context环境变量中清理traceInfo
 */
func (t *XTraceClient) CleanTraceInfoLite(irisCtx iris.Context) {
	t.cleanTraceInfoCommon(irisCtx, false)
}

/**
* 从Context环境变量中清理traceInfo
 */
func (t *XTraceClient) cleanTraceInfoCommon(irisCtx iris.Context, fullClean bool) {
	if irisCtx == nil || !t.ByResponse {
		return
	}
	if len(t.KeyTraceId) > 0 && fullClean {
		irisCtx.ResponseWriter().Header().Del(t.KeyTraceId)
	}
	if len(t.KeyTs) > 0 {
		irisCtx.ResponseWriter().Header().Del(t.KeyTs)
	}
	if len(t.KeyUri) > 0 {
		irisCtx.ResponseWriter().Header().Del(t.KeyUri)
	}
	if len(t.KeyUserId) > 0 && fullClean {
		irisCtx.ResponseWriter().Header().Del(t.KeyUserId)
	}
	if len(t.KeyUserName) > 0 && fullClean {
		irisCtx.ResponseWriter().Header().Del(t.KeyUserName)
	}
	if len(t.KeyRoleId) > 0 && fullClean {
		irisCtx.ResponseWriter().Header().Del(t.KeyRoleId)
	}
}

/**
* 从Context环境变量中获取traceId
 */
func (t *XTraceClient) TraceIdFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	pTraceInfoAny := ctx.Value(t.KeyTraceSave)
	if pTraceInfoAny == nil {
		return ""
	}
	pTraceInfo := pTraceInfoAny.(*XTraceInfo)
	if pTraceInfo == nil {
		return ""
	}
	return pTraceInfo.TraceId
}

/**
* 从Context环境变量中获取URI
 */
func (t *XTraceClient) URIFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	pTraceInfoAny := ctx.Value(t.KeyTraceSave)
	if pTraceInfoAny == nil {
		return ""
	}
	pTraceInfo := pTraceInfoAny.(*XTraceInfo)
	if pTraceInfo == nil {
		return ""
	}
	return pTraceInfo.Uri
}

/**
* 从Context环境变量中获取userName
 */
func (t *XTraceClient) UserNameFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	pTraceInfoAny := ctx.Value(t.KeyTraceSave)
	if pTraceInfoAny == nil {
		return ""
	}
	pTraceInfo := pTraceInfoAny.(*XTraceInfo)
	if pTraceInfo == nil {
		return ""
	}
	return pTraceInfo.UserName
}

/**
* 从Context环境变量中获取userId
 */
func (t *XTraceClient) UserIdFromContext(ctx context.Context) uint {
	if ctx == nil {
		return 0
	}
	pTraceInfoAny := ctx.Value(t.KeyTraceSave)
	if pTraceInfoAny == nil {
		return 0
	}
	pTraceInfo := pTraceInfoAny.(*XTraceInfo)
	if pTraceInfo == nil {
		return 0
	}
	return pTraceInfo.UserId
}

/**
* 从Context环境变量中获取roleId
 */
func (t *XTraceClient) RoleIdFromContext(ctx context.Context) uint32 {
	if ctx == nil {
		return 0
	}
	pTraceInfoAny := ctx.Value(t.KeyTraceSave)
	if pTraceInfoAny == nil {
		return 0
	}
	pTraceInfo := pTraceInfoAny.(*XTraceInfo)
	if pTraceInfo == nil {
		return 0
	}
	return pTraceInfo.RoleId
}

/*
* 时间统计插入点
ctx 环境
eventName 事件名称或者步骤名称
*/
func (t *XTraceClient) TraceTimePoint(ctx context.Context, eventName string) {
	if ctx == nil {
		return
	}
	pTraceInfoAny := ctx.Value(t.KeyTraceSave)
	if pTraceInfoAny == nil {
		return
	}
	pTraceInfo := pTraceInfoAny.(*XTraceInfo)

	if pTraceInfo == nil {
		return
	}
	lenTs := len(pTraceInfo.TsArr)
	lenEvent := len(pTraceInfo.Events)
	if lenTs <= 0 || lenEvent != lenTs-1 {
		return
	}
	pTraceInfo.TsArr = append(pTraceInfo.TsArr, time.Now().UnixMilli())
	var event string
	if len(eventName) <= 0 {
		event = "step" + strconv.Itoa(lenTs-1)
	} else {
		event = eventName
	}
	pTraceInfo.Events = append(pTraceInfo.Events, event)
	return
}

/*
* 时间统计打印信息
ctx 环境
*/
func (t *XTraceClient) TraceTimePrint(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	pTraceInfoAny := ctx.Value(t.KeyTraceSave)
	if pTraceInfoAny == nil {
		return ""
	}
	pTraceInfo := pTraceInfoAny.(*XTraceInfo)

	if pTraceInfo == nil {
		return ""
	}
	lenTs := len(pTraceInfo.TsArr)
	lenEvent := len(pTraceInfo.Events)
	if lenTs <= 0 || lenEvent != lenTs-1 {
		return ""
	}
	tsNow := time.Now().UnixMilli()
	var build strings.Builder
	//build.WriteString("traceId：")
	//build.WriteString(pTraceInfo.TraceId)
	//build.WriteString("，")
	build.WriteString("耗时统计(毫秒)，总计耗时：")
	build.WriteString(strconv.FormatInt(tsNow-pTraceInfo.TsArr[0], 10))

	if lenEvent > 0 {
		build.WriteString("，分步耗时：")
		for i := 0; i < lenEvent; i++ {
			build.WriteString(pTraceInfo.Events[i])
			build.WriteString("：")
			build.WriteString(strconv.FormatInt(pTraceInfo.TsArr[i+1]-pTraceInfo.TsArr[i], 10))
			build.WriteString("，")
		}
		build.WriteString("setp-end：")
		build.WriteString(strconv.FormatInt(tsNow-pTraceInfo.TsArr[lenEvent], 10))
		build.WriteString("。")
	} else {
		build.WriteString("。")
	}
	return build.String()
}
