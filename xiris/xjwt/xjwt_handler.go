package xjwt

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/ruomm/goxframework/gox/corex"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
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
	TraceTsKey string
	Uri        string //uri相对路径
	//UriAbs       string   //uri绝对路径
	UriFull      string   //URI全路径
	AuthToken    string   // 授权信息
	ApiPass      bool     // Api是否匹配到
	Method_Pass  bool     //方法是否pass
	IpAllow_Pass bool     //IP限制是否通过
	Mode         AuthMode // 0.不需要授权 1.需要强制授权 2.不需要强制授权，可以在业务逻辑中区分 9.业务拒绝
	Tag          string   //# tag标识，使用tag标识可以区分不同角色权限控制
}

type XjwAuthHander func(irisCtx iris.Context, verifyResult *ApiVerfiyResult) bool
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
func (t *XjwtHander) LoadConfigOpen(contextPath string, traceIdGetKey string, traceIdWriteKey string, traceTsKey string, uriHeaderKey string, toResponseHeader bool) error {
	webAPiConfigsByYaml := WebApiConfigs{}
	webAPiConfigsByYaml.OpenMode = true
	if contextPath != "" {
		webAPiConfigsByYaml.WebContextPath = contextPath
	}
	if traceIdGetKey != "" {
		webAPiConfigsByYaml.TraceIdGetKey = traceIdGetKey
	}
	if traceIdWriteKey != "" {
		webAPiConfigsByYaml.TraceIdKey = traceIdWriteKey
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
func (t *XjwtHander) LoadConfigByYaml(confYaml string, contextPath string, traceIdGetKey string, traceIdKey string, traceTsKey string, uriHeaderKey string, toResponseHeader bool) error {
	byteData, err := ioutil.ReadFile(getAbsDir(confYaml))
	if err != nil {
		panic(fmt.Sprintf("Web XjwtHander load config form yaml file err:%v", err))
		//return err
	}
	webAPiConfigsByYaml := WebApiConfigs{}
	err = yaml.Unmarshal(byteData, &webAPiConfigsByYaml)
	if err != nil {
		panic(fmt.Sprintf("Web XjwtHander read config yaml to WebAPiConfigs err:%v", err))
		//return err
	}
	webAPiConfigsByYaml.OpenMode = false
	if contextPath != "" {
		webAPiConfigsByYaml.WebContextPath = contextPath
	}
	if traceIdGetKey != "" {
		webAPiConfigsByYaml.TraceIdGetKey = traceIdGetKey
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
				//return err
			} else if len(absUri) <= 0 {
				panic(fmt.Sprintf("Web XjwtHander Parse URI Error:%v", err))
				//return err
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
func (t *XjwtHander) GetJWTHandler() func(irisCtx iris.Context) {
	handler := func(irisCtx iris.Context) {
		reqMethod := irisCtx.Method()
		fullURI := irisCtx.FullRequestURI()
		// 判断traceIdKey不为空获取traceID
		if t.webApiConfigs.TraceIdGetKey != "" {
			traceId := irisCtx.GetHeader(t.webApiConfigs.TraceIdGetKey)
			if len(traceId) <= 0 {
				timeNow := time.Now()
				traceId = "be-" + corex.TimeFormatByString("20060102-150405", &timeNow) + "-" + fmt.Sprintf("%06d", rand.Intn(1000000))
				if t.webApiConfigs.ToResponseHeader {
					irisCtx.Header(t.webApiConfigs.TraceIdKey, traceId)
				} else {
					irisCtx.Request().Header.Set(t.webApiConfigs.TraceIdKey, traceId)
				}
			} else {
				if t.webApiConfigs.ToResponseHeader {
					irisCtx.Header(t.webApiConfigs.TraceIdKey, traceId)
				} else {
					if t.webApiConfigs.TraceIdKey != t.webApiConfigs.TraceIdGetKey {
						irisCtx.Request().Header.Set(t.webApiConfigs.TraceIdKey, traceId)
					}
				}
			}
		}
		if t.webApiConfigs.TraceTsKey != "" {
			if t.webApiConfigs.ToResponseHeader {
				irisCtx.Header(t.webApiConfigs.TraceTsKey, strconv.FormatInt(time.Now().UnixMilli(), 10))
			} else {
				irisCtx.Request().Header.Set(t.webApiConfigs.TraceTsKey, strconv.FormatInt(time.Now().UnixMilli(), 10))
			}
		}
		fullRelativeURI := t.getFullRelativeURI(irisCtx, fullURI)
		if t.webApiConfigs.UriHeaderKey != "" && len(fullURI) > 0 {
			if t.webApiConfigs.ToResponseHeader {
				irisCtx.Header(t.webApiConfigs.UriHeaderKey, fullRelativeURI)
			} else {
				irisCtx.Request().Header.Set(t.webApiConfigs.UriHeaderKey, fullRelativeURI)
			}
		}
		authPassResult := false
		if t.webApiConfigs.OpenMode {
			authPassResult = true
		} else {
			var apiVerifyResult *ApiVerfiyResult = nil
			for _, webUriConfig := range t.webApiConfigs.WebApiConfigs {
				tmpVerifyResult := t.verifyApiConfig(irisCtx, fullURI, fullRelativeURI, reqMethod, &webUriConfig)
				if nil != tmpVerifyResult {
					apiVerifyResult = tmpVerifyResult
					break
				}
			}
			if nil == apiVerifyResult {
				//gobalUriAbs, _ := parseUrlWithContextPath(t.webApiConfigs.WebContextPath, "")
				gobalMode := parseAuthMode(t.webApiConfigs.DefaultMode)
				gobalAuthToken := ""
				if gobalMode == AUTH_MAY || gobalMode == AUTH_FORCE {
					gobalAuthToken = getAccessToken(irisCtx, t.webApiConfigs.CookieAuthKey, t.webApiConfigs.HeaderAuthKey, t.webApiConfigs.UrlParamAuthKey, t.webApiConfigs.AuthInfoHeader)
				}
				gobalApiVerifyResult := ApiVerfiyResult{
					TraceIdKey: t.webApiConfigs.TraceIdKey,
					TraceTsKey: t.webApiConfigs.TraceTsKey,
					Uri:        "",
					//UriAbs:     gobalUriAbs,
					UriFull:   fullRelativeURI,
					ApiPass:   false,
					Mode:      gobalMode,
					AuthToken: gobalAuthToken,
				}
				apiVerifyResult = &gobalApiVerifyResult
			}

			if apiVerifyResult.Mode == AUTH_OPEN {
				authPassResult = true
			}
			if nil != t.authHander {
				authPassResult = t.authHander(irisCtx, apiVerifyResult)
			}
		}

		//传递到后续处理器
		for {
			h := irisCtx.NextHandler()
			if h != nil {
				if authPassResult {
					irisCtx.Next()
				} else {
					//校验失败
					handleName := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
					if strings.HasPrefix(handleName, "github.com/kataras/iris") {
						//跳过业务处理
						irisCtx.Skip()
					} else {
						irisCtx.Next()
					}
				}
			} else {
				break
			}
		}
	}
	return handler
}

func (t *XjwtHander) getFullRelativeURI(irisCtx iris.Context, fullURI string) string {
	relativeURI := ""
	absURI := irisCtx.AbsoluteURI("/")
	if len(fullURI) <= 0 || len(fullURI) <= len(absURI)+1 {
		relativeURI = fullURI
	} else if !strings.HasPrefix(fullURI, absURI) {
		relativeURI = fullURI
	} else {
		// 获取接口的相对路径
		relativeURI = fullURI[len(absURI):]
	}
	return relativeURI
}

// 判断接口是否URI匹配和token校验是否通过
func (t *XjwtHander) verifyApiConfig(irisCtx iris.Context, fullURI string, fullRelativeURI string, reqMethod string, webUriConfig *WebUriConfig) *ApiVerfiyResult {
	//获取绝对路径
	absURI := irisCtx.AbsoluteURI(webUriConfig.UriAbs)
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
	ipRemoteAddr := getIpRemoteAddr(irisCtx, webUriConfig.AgentRealIpHeader)
	ipAllowGobal := verfiyIPAllow(ipRemoteAddr, webUriConfig.IpAllow)
	//if !ipAllowGobal {
	//	return &ApiVerfiyResult{TraceIdKey: t.webApiConfigs.TraceIdKey, TraceTsKey: t.webApiConfigs.TraceTsKey, Uri: webUriConfig.Uri, UriFull: fullRelativeURI, ApiPass: false, Method_Pass: false, IpAllow_Pass: false, Mode: AUTH_REFUSE}
	//}
	// 判断规则是否通过
	apiVerfiyResult := ApiVerfiyResult{TraceIdKey: t.webApiConfigs.TraceIdKey, TraceTsKey: t.webApiConfigs.TraceTsKey, Uri: webUriConfig.Uri, UriFull: fullRelativeURI, ApiPass: false, Method_Pass: false, IpAllow_Pass: false, Mode: parseAuthMode(webUriConfig.DefaultMode)}
	for _, apiConfig := range *(webUriConfig.ApiConfigs) {
		if corex.MatchStringCommon(apiConfig.Api, relativeURI) {
			apiVerfiyResult.ApiPass = true
			apiVerfiyResult.Method_Pass = verifyRequestMethod(apiConfig.Method, reqMethod)
			apiVerfiyResult.IpAllow_Pass = verfiyIPAllow(ipRemoteAddr, apiConfig.IpAllow)
			apiVerfiyResult.Mode = parseAuthMode(apiConfig.Mode)
			apiVerfiyResult.Tag = apiConfig.Tag
			if apiVerfiyResult.Method_Pass {
				break
			}
		}
	}
	// 校验IP和API方法
	if !ipAllowGobal || !apiVerfiyResult.IpAllow_Pass {
		// 如是IP校验未通过则设置为拒绝
		apiVerfiyResult.Mode = AUTH_REFUSE
	} else if !apiVerfiyResult.ApiPass || !apiVerfiyResult.Method_Pass {
		// 如是方法没有通过，则使用默认的处理
		apiVerfiyResult.Mode = parseAuthMode(webUriConfig.DefaultMode)
	}
	if apiVerfiyResult.Mode == AUTH_FORCE || apiVerfiyResult.Mode == AUTH_MAY {
		apiVerfiyResult.AuthToken = getAccessToken(irisCtx, webUriConfig.CookieAuthKey, webUriConfig.HeaderAuthKey, t.webApiConfigs.UrlParamAuthKey, webUriConfig.AuthInfoHeader)
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
func getIpRemoteAddr(irisCtx iris.Context, agentRealIpHeader string) string {
	ipRemoteAddr := irisCtx.GetHeader(agentRealIpHeader)
	if ipRemoteAddr == "" {
		ipRemoteAddr = irisCtx.RemoteAddr()
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
		fmt.Println("Web XjwtHander ipAllow Invalid CIDR err:" + err.Error())
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
func getAccessToken(irisCtx iris.Context, cookieAuthKey, headerAuthKey, uriParamAuthKey string, authInfoHeader string) string {
	accessToken := ""
	// 尝试从Cookie里面取用户授权凭证
	if cookieAuthKey != "" {
		pCookie, err := irisCtx.GetRequestCookie(cookieAuthKey)
		if err == nil {
			accessToken = pCookie.Value
		}
	}
	// 尝试从header里面取用户授权凭证
	if len(accessToken) <= 8 {
		if headerAuthKey != "" {
			accessToken = irisCtx.GetHeader(headerAuthKey)
		}
	}
	// 尝试从请求参数里面取用户授权凭证
	if len(accessToken) <= 8 {
		if uriParamAuthKey != "" {
			accessToken = xJwtGetUrlToken(irisCtx, uriParamAuthKey)
		}

	}
	// 去除校验头
	if authInfoHeader != "" && accessToken != "" {
		if strings.HasPrefix(accessToken, authInfoHeader+" ") {
			accessToken = accessToken[len(authInfoHeader)+1:]
		} else if strings.HasPrefix(accessToken, authInfoHeader+":") {
			accessToken = accessToken[len(authInfoHeader)+1:]
		} else if strings.HasPrefix(accessToken, authInfoHeader) {
			accessToken = accessToken[len(authInfoHeader):]
		}
	}
	return accessToken
}
func xJwtGetUrlToken(irisCtx iris.Context, origKey string) string {
	tmpToken := xJwtGetUrlQuery(irisCtx, origKey)
	if len(tmpToken) <= 8 {
		tmpToken = xJwtGetUrlParam(irisCtx, origKey)
	}
	return tmpToken
}

func xJwtGetUrlQuery(irisCtx iris.Context, origKey string) string {
	if len(origKey) <= 0 {
		return ""
	}
	if irisCtx.URLParamExists(origKey) {
		paramVal := irisCtx.URLParam(origKey)
		return paramVal
	} else {
		return ""
	}
}

func xJwtGetUrlParam(irisCtx iris.Context, origKey string) string {
	if len(origKey) <= 0 {
		return ""
	}
	if irisCtx.Params().Exists(origKey) {
		paramVal := irisCtx.Params().GetString(origKey)
		return paramVal
	} else {
		return ""
	}
}

// 获取用户授权凭证信息
//func getAccessToken(irisCtx iris.Context, webApiConfig *WebApiConfigs) string {
//	accessToken := ""
//	// 尝试从Cookie里面取用户授权凭证
//	if webUriConfig.CookieAuthKey != "" {
//		pCookie, err := irisCtx.GetRequestCookie(webUriConfig.CookieAuthKey)
//		if err == nil {
//			accessToken = pCookie.Value
//		}
//	}
//	// 尝试从header里面取用户授权凭证
//	if len(accessToken) <= 8 {
//		if webUriConfig.HeaderAuthKey != "" {
//			accessToken = irisCtx.GetHeader(webUriConfig.HeaderAuthKey)
//		}
//	}
//	// 尝试从请求参数里面取用户授权凭证
//	//if len(accessToken) <= 8 {
//	//	tmpAccessToken := irisCtx.URLParam(common.COOKIE_KEY_AUTH_TOKEN)
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
