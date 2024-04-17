package xjwt

type WebApiConfigs struct {
	// 是否公开模式，公开模式在不进行授权验证，只进行Header信息存储
	OpenMode bool
	//#web服务器的ContextPath，如是有值，则所有接口前缀加上ContextPath
	WebContextPath string //`yaml:"webContextPath"`
	//# 路径追踪的traceID,如需要日志路径追踪则需要提供
	TraceIdKey string //`yaml:"traceIdKey"`
	//# 毫秒时间存储的key，提供此值则往header的此key里面写入此毫秒时间戳，可以方便做时间切片
	TraceTsKey string //`yaml:"traceTsKey"`
	//# uri存储的key，提供此值则往header的此key里面写入此URI，可以方便做日志输出
	UriHeaderKey string //`yaml:"uriHeaderKey"`
	// # 是否存储到response的header里面
	ToResponseHeader bool
	//# 授权信息存储的字段
	HeaderAuthKey string `yaml:"headerAuthKey"`
	//# 授权信息在cookie里面存储的字段
	CookieAuthKey string `yaml:"cookieAuthKey"`
	//# 授权信息开头字段，部分规范接口中会有Bearer开头的字段，需要截取掉
	AuthInfoHeader string `yaml:"authInfoHeader"`
	//# 代理转发服务的真实IP写入字段
	AgentRealIpHeader string `yaml:"agentRealIpHeader"`
	//# 没有匹配到任何API分组(WebUriConfig下面的uri路径)时候的时候的授权策略，建议refuse、force。
	DefaultMode string `yaml:"defaultMode"`
	//# API授权配置
	WebApiConfigs []WebUriConfig `yaml:"webApiConfigs"`
}

// # API授权配置，apiConfigs是根节点
// # api、method,mode组成授权标识，是apiConfigs的节点
type WebUriConfig struct {
	//# 授权信息存储的字段
	HeaderAuthKey string `yaml:"headerAuthKey"`
	//# 授权信息在cookie里面存储的字段
	CookieAuthKey string `yaml:"cookieAuthKey"`
	//# 授权信息开头字段，部分规范接口中会有Bearer开头的字段，需要截取掉
	AuthInfoHeader string `yaml:"authInfoHeader"`
	//# 代理转发服务的真实IP写入字段
	AgentRealIpHeader string `yaml:"agentRealIpHeader"`
	//# API分组，以uri来区分不同的组别，uri以/开头则是绝对路径，不以/开头则是相对路径，会拼接上webContextPath的前缀路径
	Uri string `yaml:"uri"`
	//# API分组的绝对API路径
	UriAbs string
	//# 匹配到rui(API分组)但没有匹配api时候的授权策略，建议refuse、force。
	DefaultMode string `yaml:"defaultMode"`
	// # ipAllow 限制IP访问的范围，如192.168.3.1,10.0.1.1/16,192.168.50.3-192.168.50.6等。
	IpAllow string `yaml:"ipAllow"`
	//# apiConfigs配置详细API的授权策略
	ApiConfigs *[]ApiConfig `yaml:"apiConfigs"`
}

// # apiConfigs配置详细API的授权策略
type ApiConfig struct {
	//# apiConfigs配置详细API的授权策略
	//# api是rui下面的子路径。
	//# api子路径以^开头或$结尾则匹配正则表达式
	//# 非正则表示的api子路径支持开头*和结尾*通配符
	Api string `yaml:"api"`
	//# method代表授权的http请求方法，ALL表示http请求方法，方法以http标准需要大写
	Method string `yaml:"method"`
	//# mode 代表授权策略，open：开放接口，不需要授权，force：需要强制授权，may：授权和不授权都可以，refuse：拒绝执行
	Mode string `yaml:"mode"`
	//# tag标识，使用tag标识可以区分不同角色权限控制
	Tag string `yaml:"tag"`
	// # ipAllow 限制IP访问的范围，如192.168.3.1,10.0.1.1/16,192.168.50.3-192.168.50.6等。
	IpAllow string `yaml:"ipAllow"`
}
