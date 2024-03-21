package xjwt

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/ruomm/goxframework/gox/corex"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type AuthMode string

const (
	// 不需要授权
	AUTH_OPEN AuthMode = "open"
	// 需要强制授权
	AUTH_FORCE AuthMode = "force"
	// 不需要强制授权，可以在业务逻辑中区分
	AUTH_MAY AuthMode = "may"
	// AUTH_REFUSE
	AUTH_REFUSE AuthMode = "refuse"
)

type ApiVerfiyResult struct {
	//# 路径追踪的traceID,如需要日志路径追踪则需要提供
	TraceIdKey string
	//# 毫秒时间存储的key，提供此值则往header的此key里面写入此毫秒时间戳，可以方便做时间切片
	TraceTsKey   string
	Uri          string   //uri相对路径
	UriAbs       string   //uri绝对路径
	AuthToken    string   // 授权信息
	ApiPass      bool     // Api是否匹配到
	Method_Pass  bool     //方法是否pass
	IpAllow_Pass bool     //IP限制是否通过
	Mode         AuthMode // 0.不需要授权 1.需要强制授权 2.不需要强制授权，可以在业务逻辑中区分 9.业务拒绝
}

type XjwAuthHander func(ctx iris.Context, verifyResult *ApiVerfiyResult) bool
type XjwtHander struct {
	config        string
	webApiConfigs *WebApiConfigs
	authHander    XjwAuthHander
}

func (t *XjwtHander) UseAuthHander(authHander XjwAuthHander) {
	t.authHander = authHander
}

// 不加载API接口配置文件，只做traceId，TraceTs缓存
/**
contextPath：服务器URI根路径
traceIdKey：traceId的headerKey
traceTsKey：traceTs的headerKey
uriHeaderKey：uri的headerKey
toResponseHeader：存储header信息是否到response的header里面
*/
func (t *XjwtHander) LoadConfigOpen(contextPath string, traceIdKey string, traceTsKey string, uriHeaderKey string, toResponseHeader bool) error {
	webAPiConfigsByYaml := WebApiConfigs{}
	webAPiConfigsByYaml.OpenMode = true
	if contextPath != "" {
		webAPiConfigsByYaml.WebContextPath = contextPath
	}
	if traceIdKey != "" {
		webAPiConfigsByYaml.TraceIdKey = traceIdKey
	}
	if traceTsKey != "" {
		webAPiConfigsByYaml.TraceTsKey = traceTsKey
	}
	if uriHeaderKey != "" {
		webAPiConfigsByYaml.UriHeaderKey = uriHeaderKey
	}
	webAPiConfigsByYaml.ToResponseHeader = toResponseHeader
	t.webApiConfigs = &webAPiConfigsByYaml
	return nil
}

// 加载API接口配置文件
/**
confYaml:配置文件路径
contextPath：服务器URI根路径
traceIdKey：traceId的headerKey
traceTsKey：traceTs的headerKey
uriHeaderKey：uri的headerKey
toResponseHeader：存储header信息是否到response的header里面
*/
func (t *XjwtHander) LoadConfigByYaml(confYaml string, contextPath string, traceIdKey string, traceTsKey string, uriHeaderKey string, toResponseHeader bool) error {
	byteData, err := ioutil.ReadFile(getAbsDir(confYaml))
	if err != nil {
		panic(fmt.Sprintf("Web XjwtHander load config form yaml file err:%v", err))
		return err
	}
	webAPiConfigsByYaml := WebApiConfigs{}
	err = yaml.Unmarshal(byteData, &webAPiConfigsByYaml)
	if err != nil {
		panic(fmt.Sprintf("Web XjwtHander read config yaml to WebAPiConfigs err:%v", err))
		return err
	}
	webAPiConfigsByYaml.OpenMode = false
	if contextPath != "" {
		webAPiConfigsByYaml.WebContextPath = contextPath
	}
	if traceIdKey != "" {
		webAPiConfigsByYaml.TraceIdKey = traceIdKey
	}
	if traceTsKey != "" {
		webAPiConfigsByYaml.TraceTsKey = traceTsKey
	}
	if uriHeaderKey != "" {
		webAPiConfigsByYaml.UriHeaderKey = uriHeaderKey
	}
	webAPiConfigsByYaml.ToResponseHeader = toResponseHeader
	for i := 0; i < len(webAPiConfigsByYaml.WebApiConfigs); i++ {
		webAPiConfig := webAPiConfigsByYaml.WebApiConfigs[i]
		absUri, errUri := parseUrlWithContextPath(webAPiConfigsByYaml.WebContextPath, webAPiConfig.Uri)
		if errUri != nil {
			if err != nil {
				panic(fmt.Sprintf("Web XjwtHander Parse URI Error:%v", err))
				return err
			} else if len(absUri) <= 0 {
				panic(fmt.Sprintf("Web XjwtHander Parse URI Error:%v", err))
				return err
			}
		}
		webAPiConfigsByYaml.WebApiConfigs[i].UriAbs = absUri
		if webAPiConfigsByYaml.WebApiConfigs[i].HeaderAuthKey == "" {
			webAPiConfigsByYaml.WebApiConfigs[i].HeaderAuthKey = webAPiConfigsByYaml.HeaderAuthKey
		}
		if webAPiConfigsByYaml.WebApiConfigs[i].CookieAuthKey == "" {
			webAPiConfigsByYaml.WebApiConfigs[i].CookieAuthKey = webAPiConfigsByYaml.CookieAuthKey
		}
		if webAPiConfigsByYaml.WebApiConfigs[i].AuthInfoHeader == "" {
			webAPiConfigsByYaml.WebApiConfigs[i].AuthInfoHeader = webAPiConfigsByYaml.AuthInfoHeader
		}
		if webAPiConfigsByYaml.WebApiConfigs[i].AgentRealIpHeader == "" {
			webAPiConfigsByYaml.WebApiConfigs[i].AgentRealIpHeader = webAPiConfigsByYaml.AgentRealIpHeader
		}
		if webAPiConfigsByYaml.WebApiConfigs[i].AgentRealIpHeader == "" {
			webAPiConfigsByYaml.WebApiConfigs[i].AgentRealIpHeader = "X-Real-IP"
		}
	}
	fmt.Println("Web XjwtHander config init success!")
	t.webApiConfigs = &webAPiConfigsByYaml
	return nil
}

//	func (t *XjwtHander) LoadConfigByYaml(confYaml string) error {
//		return t.LoadConfigByYamlWithContextPath(confYaml, "", "", "")
//	}
func (t *XjwtHander) GetJWTHandler() func(ctx iris.Context) {
	handler := func(ctx iris.Context) {
		reqMethod := ctx.Method()
		fullURI := ctx.FullRequestURI()
		// 判断traceIdKey不为空获取traceID
		if t.webApiConfigs.TraceIdKey != "" {
			traceId := ctx.GetHeader(t.webApiConfigs.TraceIdKey)
			if len(traceId) <= 0 {
				traceId = "be-" + strconv.FormatInt(time.Now().Unix(), 10)
				if t.webApiConfigs.ToResponseHeader {
					ctx.Header(t.webApiConfigs.TraceIdKey, traceId)
				} else {
					ctx.Request().Header.Set(t.webApiConfigs.TraceIdKey, traceId)
				}

			}
		}
		if t.webApiConfigs.TraceTsKey != "" {
			if t.webApiConfigs.ToResponseHeader {
				ctx.Header(t.webApiConfigs.TraceTsKey, strconv.FormatInt(time.Now().UnixMilli(), 10))
			} else {
				ctx.Request().Header.Set(t.webApiConfigs.TraceTsKey, strconv.FormatInt(time.Now().UnixMilli(), 10))
			}
		}
		if t.webApiConfigs.UriHeaderKey != "" {
			absURI := ctx.AbsoluteURI("/")
			if t.webApiConfigs.ToResponseHeader {
				ctx.Header(t.webApiConfigs.UriHeaderKey, absURI)
			} else {
				ctx.Request().Header.Set(t.webApiConfigs.UriHeaderKey, absURI)
			}
			//if len(absURI) > 0 {
			//	if t.webApiConfigs.ToResponseHeader {
			//		ctx.Header(t.webApiConfigs.UriHeaderKey, absURI)
			//	} else {
			//		ctx.Request().Header.Set(t.webApiConfigs.UriHeaderKey, absURI)
			//	}
			//}

		}
		authPassResult := false
		if t.webApiConfigs.OpenMode {
			authPassResult = true
		} else {
			var apiVerifyResult *ApiVerfiyResult = nil
			for _, webUriConfig := range t.webApiConfigs.WebApiConfigs {
				tmpVerifyResult := t.verifyApiConfig(ctx, fullURI, reqMethod, &webUriConfig)
				if nil != tmpVerifyResult {
					apiVerifyResult = tmpVerifyResult
					break
				}
			}
			if nil == apiVerifyResult {
				gobalUriAbs, _ := parseUrlWithContextPath(t.webApiConfigs.WebContextPath, "")
				gobalMode := parseAuthMode(t.webApiConfigs.DefaultMode)
				gobalAuthToken := ""
				if gobalMode == AUTH_MAY || gobalMode == AUTH_FORCE {
					gobalAuthToken = getAccessToken(ctx, t.webApiConfigs.CookieAuthKey, t.webApiConfigs.HeaderAuthKey, t.webApiConfigs.AuthInfoHeader)
				}
				gobalApiVerifyResult := ApiVerfiyResult{
					TraceIdKey: t.webApiConfigs.TraceIdKey,
					TraceTsKey: t.webApiConfigs.TraceTsKey,
					Uri:        "",
					UriAbs:     gobalUriAbs,
					ApiPass:    false,
					Mode:       gobalMode,
					AuthToken:  gobalAuthToken,
				}
				apiVerifyResult = &gobalApiVerifyResult
			}

			if apiVerifyResult.Mode == AUTH_OPEN {
				authPassResult = true
			}
			if nil != t.authHander {
				authPassResult = t.authHander(ctx, apiVerifyResult)
			}
		}

		//传递到后续处理器
		for {
			h := ctx.NextHandler()
			if h != nil {
				if authPassResult {
					ctx.Next()
				} else {
					//校验失败
					handleName := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
					if strings.HasPrefix(handleName, "github.com/kataras/iris") {
						//跳过业务处理
						ctx.Skip()
					} else {
						ctx.Next()
					}
				}
			} else {
				break
			}
		}
	}
	return handler
}

// 判断接口是否URI匹配和token校验是否通过
func (t *XjwtHander) verifyApiConfig(ctx iris.Context, fullURI string, reqMethod string, webUriConfig *WebUriConfig) *ApiVerfiyResult {
	//获取绝对路径
	absURI := ctx.AbsoluteURI(webUriConfig.UriAbs)
	if len(fullURI) <= 0 || len(fullURI) <= len(absURI)+1 || len(reqMethod) <= 0 {
		return nil
	}
	// 校验API是否属于此接口
	if !strings.HasPrefix(fullURI, absURI) {
		return nil
	}
	// 获取接口的相对路径
	relativeURI := fullURI[len(absURI)+1:]
	// 获取接口的真实地址
	ipRemoteAddr := getIpRemoteAddr(ctx, webUriConfig.AgentRealIpHeader)
	ipAllowGobal := verfiyIPAllow(ipRemoteAddr, webUriConfig.IpAllow)
	if !ipAllowGobal {
		return &ApiVerfiyResult{TraceIdKey: t.webApiConfigs.TraceIdKey, TraceTsKey: t.webApiConfigs.TraceTsKey, Uri: webUriConfig.Uri, UriAbs: webUriConfig.UriAbs, ApiPass: false, Method_Pass: false, IpAllow_Pass: false, Mode: AUTH_REFUSE}
	}
	// 判断规则是否通过
	apiVerfiyResult := ApiVerfiyResult{TraceIdKey: t.webApiConfigs.TraceIdKey, TraceTsKey: t.webApiConfigs.TraceTsKey, Uri: webUriConfig.Uri, UriAbs: webUriConfig.UriAbs, ApiPass: false, Method_Pass: false, IpAllow_Pass: false, Mode: parseAuthMode(webUriConfig.DefaultMode)}
	for _, apiConfig := range *(webUriConfig.ApiConfigs) {
		if corex.MatchStringCommon(apiConfig.Api, relativeURI) {
			apiVerfiyResult.ApiPass = true
			apiVerfiyResult.Method_Pass = verifyRequestMethod(apiConfig.Method, reqMethod)
			apiVerfiyResult.IpAllow_Pass = verfiyIPAllow(ipRemoteAddr, apiConfig.IpAllow)
			apiVerfiyResult.Mode = parseAuthMode(apiConfig.Mode)
			if apiVerfiyResult.Method_Pass {
				break
			}
		}
	}
	// 如是方法没有通过，则使用默认的处理
	if !apiVerfiyResult.ApiPass || !apiVerfiyResult.Method_Pass {
		apiVerfiyResult.Mode = parseAuthMode(webUriConfig.DefaultMode)
	}
	if !apiVerfiyResult.IpAllow_Pass {
		apiVerfiyResult.Mode = AUTH_REFUSE
	}
	if apiVerfiyResult.Mode == AUTH_FORCE || apiVerfiyResult.Mode == AUTH_MAY {
		apiVerfiyResult.AuthToken = getAccessToken(ctx, webUriConfig.CookieAuthKey, webUriConfig.HeaderAuthKey, webUriConfig.AuthInfoHeader)
	}
	return &apiVerfiyResult
}

// 验证是否严格uri和api通过，如是通过则不继续验证
func verfiyResultApiPass(t *ApiVerfiyResult) bool {
	if t.ApiPass && t.Method_Pass {
		return true
	} else {
		return false
	}
}

// 解析授权方式
func parseAuthMode(mode string) AuthMode {
	if mode == string(AUTH_OPEN) {
		return AUTH_OPEN
	} else if mode == string(AUTH_FORCE) {
		return AUTH_FORCE
	} else if mode == string(AUTH_MAY) {
		return AUTH_MAY
	} else if mode == string(AUTH_REFUSE) {
		return AUTH_REFUSE
	} else {
		return AUTH_REFUSE
	}
}
func getIpRemoteAddr(ctx iris.Context, agentRealIpHeader string) string {
	ipRemoteAddr := ctx.GetHeader(agentRealIpHeader)
	if ipRemoteAddr == "" {
		ipRemoteAddr = ctx.RemoteAddr()
	}
	return ipRemoteAddr
}

// 验证IP限制是否通过
func verfiyIPAllow(ipRemoteAddr string, ipAllow string) bool {
	if ipAllow == "" {
		return true
	}
	if ipRemoteAddr == "" {
		return false
	}
	//println("请求IP地址为：" + ipRequest)
	cidrsArr := strings.Split(ipAllow, ",")
	ipAllowPass := false
	for _, cidrs := range cidrsArr {
		if cidrs == "" {
			continue
		}
		if strings.Contains(cidrs, "/") {
			if isIpInRange(ipRemoteAddr, cidrs) {
				ipAllowPass = true
				break
			}
		} else if strings.Contains(cidrs, "-") {
			if isIpInStartEnd(ipRemoteAddr, cidrs) {
				ipAllowPass = true
				break
			}
		} else if strings.Compare(ipRemoteAddr, cidrs) == 0 {
			ipAllowPass = true
			break
		}
	}
	return ipAllowPass
}
func isIpInRange(ip string, cidr string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		fmt.Println("Web XjwtHander ipAllow Invalid CIDR err:%v", err)
		return false
	}
	parsedIp := net.ParseIP(ip)
	return ipNet.Contains(parsedIp)
}
func isIpInStartEnd(ip string, ipAreaStr string) bool {
	ipAreaS := strings.Split(ipAreaStr, "-")
	if len(ipAreaS) != 2 {
		return false
	}
	if ipAreaS[0] == "" || ipAreaS[1] == "" {
		return false
	}
	if strings.Compare(ip, ipAreaS[0]) >= 0 && strings.Compare(ip, ipAreaS[1]) <= 0 {
		return true
	} else {
		return false
	}
}

// 验证方式是否匹配
func verifyRequestMethod(configMethod, reqMethod string) bool {
	isPass := false
	if len(configMethod) <= 0 || configMethod == "ALL" {
		isPass = true
	} else {
		configMethods := strings.Split(configMethod, ",")
		for m := 0; m < len(configMethods); m++ {
			if configMethods[m] == reqMethod {
				isPass = true
				break
			}
		}
	}
	return isPass
}

// 获取用户授权凭证信息
func getAccessToken(ctx iris.Context, cookieAuthKey, headerAuthKey, authInfoHeader string) string {
	accessToken := ""
	// 尝试从Cookie里面取用户授权凭证
	if cookieAuthKey != "" {
		pCookie, err := ctx.GetRequestCookie(cookieAuthKey)
		if err == nil {
			accessToken = pCookie.Value
		}
	}
	// 尝试从header里面取用户授权凭证
	if len(accessToken) <= 8 {
		if headerAuthKey != "" {
			accessToken = ctx.GetHeader(headerAuthKey)
		}
	}
	// 尝试从请求参数里面取用户授权凭证
	//if len(accessToken) <= 8 {
	//	tmpAccessToken := ctx.URLParam(common.COOKIE_KEY_AUTH_TOKEN)
	//	if len(tmpAccessToken) > 8 {
	//		accessToken = tmpAccessToken
	//	}
	//}
	// 去除校验头
	if authInfoHeader != "" && accessToken != "" {
		if strings.HasPrefix(accessToken, authInfoHeader) {
			accessToken = accessToken[len(authInfoHeader):]
		}
	}
	return accessToken
}

// 获取用户授权凭证信息
//func getAccessToken(ctx iris.Context, webApiConfig *WebApiConfigs) string {
//	accessToken := ""
//	// 尝试从Cookie里面取用户授权凭证
//	if webUriConfig.CookieAuthKey != "" {
//		pCookie, err := ctx.GetRequestCookie(webUriConfig.CookieAuthKey)
//		if err == nil {
//			accessToken = pCookie.Value
//		}
//	}
//	// 尝试从header里面取用户授权凭证
//	if len(accessToken) <= 8 {
//		if webUriConfig.HeaderAuthKey != "" {
//			accessToken = ctx.GetHeader(webUriConfig.HeaderAuthKey)
//		}
//	}
//	// 尝试从请求参数里面取用户授权凭证
//	//if len(accessToken) <= 8 {
//	//	tmpAccessToken := ctx.URLParam(common.COOKIE_KEY_AUTH_TOKEN)
//	//	if len(tmpAccessToken) > 8 {
//	//		accessToken = tmpAccessToken
//	//	}
//	//}
//	// 去除校验头
//	if webUriConfig.AuthInfoHeader != "" && accessToken != "" {
//		if strings.HasPrefix(accessToken, webUriConfig.AuthInfoHeader) {
//			accessToken = accessToken[len(webUriConfig.AuthInfoHeader):]
//		}
//	}
//	return accessToken
//}
