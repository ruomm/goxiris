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
	ContextKeyTrace     string
	HeaderKeyTraceId    string
	HeaderKeyTs         string
	HeaderKeyAuthUserId string
	HeaderKeyUri        string
	ByResponseHeader    bool
}

/*
* keyTrace：trace信息存储key
headerKeyTraceId：traceId在header中的KEY
headerKeyTs：ts在header中的KEY
headerKeyUri：uri在header中的KEY
headerKeyAuthUserId：userId在header中的KEY
byResponseHeader：是否从responseHeader里面读取
*/
func GenXTraceClient(keyTrace string, headerKeyTraceId string, headerKeyTs string, headerKeyUri string, headerKeyAuthUserId string, byResponseHeader bool) XTraceClient {
	realKeyTrace := keyTrace
	if len(realKeyTrace) <= 0 {
		realKeyTrace = "KEY_TRACE"
	}
	realHeaderTraceId := headerKeyTraceId
	if len(realHeaderTraceId) <= 0 {
		realHeaderTraceId = "X-IDR-TraceId"
	}
	realHeaderTs := headerKeyTs
	if len(realHeaderTs) <= 0 {
		realHeaderTs = "X-IDR-Ts"
	}
	realHeaderUri := headerKeyUri
	if len(realHeaderUri) <= 0 {
		realHeaderUri = "X-IDR-Uri"
	}
	realHeaderUserId := headerKeyAuthUserId
	if len(realHeaderUserId) <= 0 {
		realHeaderUserId = "__auth_user_id"
	}
	xTraceClient := XTraceClient{
		ContextKeyTrace:     realKeyTrace,
		HeaderKeyTraceId:    realHeaderTraceId,
		HeaderKeyTs:         realHeaderTs,
		HeaderKeyAuthUserId: realHeaderUserId,
		HeaderKeyUri:        realHeaderUri,
		ByResponseHeader:    byResponseHeader,
	}
	return xTraceClient
}

type XTraceInfo struct {
	TraceId string
	TsArr   []int64
	Events  []string
	Uri     string
	UserId  uint
}

/*
* 获取一个带TraceInfo的Context环境变量
 */
func (t *XTraceClient) ToTraceContext(uCtx iris.Context) *context.Context {
	traceId := ""
	userIdStr := ""
	uriStr := ""
	tsHeader := ""
	if t.ByResponseHeader {
		traceId = uCtx.ResponseWriter().Header().Get(t.HeaderKeyTraceId)
		userIdStr = uCtx.ResponseWriter().Header().Get(t.HeaderKeyAuthUserId)
		uriStr = uCtx.ResponseWriter().Header().Get(t.HeaderKeyUri)
		tsHeader = uCtx.ResponseWriter().Header().Get(t.HeaderKeyTs)
	} else {
		traceId = uCtx.GetHeader(t.HeaderKeyTraceId)
		userIdStr = uCtx.GetHeader(t.HeaderKeyAuthUserId)
		uriStr = uCtx.GetHeader(t.HeaderKeyUri)
		tsHeader = uCtx.GetHeader(t.HeaderKeyTs)
	}
	var userId uint = 0
	if len(userIdStr) > 0 {
		tmpUserId, err := strconv.ParseUint(userIdStr, 10, 64)
		if err == nil {
			userId = uint(tmpUserId)
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
		TraceId: traceId,
		TsArr:   tsArr,
		Events:  events,
		Uri:     uriStr,
		UserId:  userId,
	}
	ctx := context.WithValue(context.Background(), t.ContextKeyTrace, &xTraceInfo)
	return &ctx
}

/**
* 从Context环境变量中获取traceId
 */
func (t *XTraceClient) TraceIdFromContext(pCtx *context.Context) string {
	if pCtx == nil {
		return ""
	}
	pTraceInfoAny := (*pCtx).Value(t.ContextKeyTrace)
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
* 从Context环境变量中获取uri
 */
func (t *XTraceClient) URIFromContext(pCtx *context.Context) string {
	if pCtx == nil {
		return ""
	}
	pTraceInfoAny := (*pCtx).Value(t.ContextKeyTrace)
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
* 从Context环境变量中获取userId
 */
func (t *XTraceClient) UserIdFromContext(pCtx *context.Context) uint {
	if pCtx == nil {
		return 0
	}
	pTraceInfoAny := (*pCtx).Value(t.ContextKeyTrace)
	if pTraceInfoAny == nil {
		return 0
	}
	pTraceInfo := pTraceInfoAny.(*XTraceInfo)
	if pTraceInfo == nil {
		return 0
	}
	return pTraceInfo.UserId
}

/*
* 时间统计插入点
uCtx 环境
eventName 事件名称或者步骤名称
*/
func (t *XTraceClient) TraceTimePoint(pCtx *context.Context, eventName string) {
	if pCtx == nil {
		return
	}
	pTraceInfoAny := (*pCtx).Value(t.ContextKeyTrace)
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
uCtx 环境
*/
func (t *XTraceClient) TraceTimePrint(pCtx *context.Context) string {
	if pCtx == nil {
		return ""
	}
	pTraceInfoAny := (*pCtx).Value(t.ContextKeyTrace)
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
