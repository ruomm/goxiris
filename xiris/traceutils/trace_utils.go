package traceutils

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/ruomm/goxiris/xiris/configx"
	"strconv"
	"strings"
	"time"
)

type ContextTraceInfo struct {
	TraceId string
	TsArr   []int64
	Events  []string
	UserId  uint
}

/*
* 获取一个带TraceInfo的Context环境变量
 */
func ToTraceContext(uCtx iris.Context) *context.Context {
	traceId := uCtx.GetHeader(configx.HEADER_NAME_TRACEID)
	userIdStr := uCtx.GetHeader(configx.HEADER_NAME_AUTH_USER_ID)
	var userId uint = 0
	if len(userIdStr) > 0 {
		tmpUserId, err := strconv.ParseUint(userIdStr, 10, 64)
		if err == nil {
			userId = uint(tmpUserId)
		}
	}
	//ts := strconv.ParseInt(uCtx.GetHeader(common.HEADER_NAME_TS), 10)
	var tsArr []int64 = nil
	var events []string = nil
	tsHeader := uCtx.GetHeader(configx.HEADER_NAME_TS)
	if tsHeader != "" {
		ts, err := strconv.ParseInt(uCtx.GetHeader(configx.HEADER_NAME_TS), 10, 64)
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
	contextTraceInfo := ContextTraceInfo{
		TraceId: traceId,
		TsArr:   tsArr,
		Events:  events,
		UserId:  userId,
	}
	ctx := context.WithValue(context.Background(), configx.CONTEXT_KEY_TRACE, &contextTraceInfo)
	return &ctx
}

/**
* 从Context环境变量中获取traceId
 */
func TraceIdFromContext(pCtx *context.Context) string {
	if pCtx == nil {
		return ""
	}
	pTraceInfoAny := (*pCtx).Value(configx.CONTEXT_KEY_TRACE)
	if pTraceInfoAny == nil {
		return ""
	}
	pTraceInfo := pTraceInfoAny.(*ContextTraceInfo)
	if pTraceInfo == nil {
		return ""
	}
	return pTraceInfo.TraceId
}

/**
* 从Context环境变量中获取userId
 */
func UserIdFromContext(pCtx *context.Context) uint {
	if pCtx == nil {
		return 0
	}
	pTraceInfoAny := (*pCtx).Value(configx.CONTEXT_KEY_TRACE)
	if pTraceInfoAny == nil {
		return 0
	}
	pTraceInfo := pTraceInfoAny.(*ContextTraceInfo)
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
func TraceTimePoint(pCtx *context.Context, eventName string) {
	if pCtx == nil {
		return
	}
	pTraceInfoAny := (*pCtx).Value(configx.CONTEXT_KEY_TRACE)
	if pTraceInfoAny == nil {
		return
	}
	pTraceInfo := pTraceInfoAny.(*ContextTraceInfo)

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
func TraceTimePrint(pCtx *context.Context) string {
	if pCtx == nil {
		return ""
	}
	pTraceInfoAny := (*pCtx).Value(configx.CONTEXT_KEY_TRACE)
	if pTraceInfoAny == nil {
		return ""
	}
	pTraceInfo := pTraceInfoAny.(*ContextTraceInfo)

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
