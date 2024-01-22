package configx

const (
	// 成功返回的默认信息
	RESP_MSG_OK      = "success"
	RESP_MSG_OK_AJAX = "request success, waiting process"
	/**
	# 响应码列表：
	        200		创建成功，通常用在同步操作时
	        202		创建成功，通常用在异步操作时，表示请求已接受，但是还没有处理完成
	        400		参数错误，通常用在表单参数错误
	        401		授权错误，通常用在 Token 缺失或失效，注意 401 会触发前端跳转到登录页
	        403		操作被拒绝，通常发生在权限不足时，注意此时务必带上详细错误信息
	        404		没有找到对象，通常发生在使用错误的 id 查询详情
	        500		服务器错误
	*/
	//创建成功，通常用在同步操作时
	ERROR_CODE_OK int = 200
	//创建成功，通常用在异步操作时，表示请求已接受，但是还没有处理完成
	ERROR_CODE_OK_AJAX int = 202
	// 参数错误，通常用在表单参数错误
	ERROR_CODE_PARAM_CHECK int = 400
	//授权错误，通常用在 Token 缺失或失效，注意 401 会触发前端跳转到登录页
	ERROR_CODE_TOKEN_INVALID int = 401
	//操作被拒绝，通常发生在权限不足时，注意此时务必带上详细错误信息
	ERROR_CODE_REFUSE int = 403
	//没有找到对象，通常发生在使用错误的 id 查询详情
	ERROR_CODE_NOT_EXIST int = 404
	//资源处理错误，通常是数据库读写错误
	ERROR_CODE_DB_CORE int = 405
	//文件读写错误
	ERROR_CODE_FILE_CORE int = 406
	//第三方授权失败
	ERROR_CODE_THRID_GATEWAY int = 411
	//服务器错误
	ERROR_CODE_INTERNAL_ERROR int = 500
	// 请求的内容无法处理，通常是不符合规范的数据格式
	ERROR_CODE_UNABLE_HANDLE int = 901
	//未定义错误
	ERROR_CODE_UNDEFINED int = 990
	//运行异常错误
	ERROR_CODE_EXCEPTION int = 999

	//agent相关
	ERROR_CODE_AGENT int = 500001
)

const (
	// Http请求头
	CONTEXT_KEY_TRACE         = "KEY_TRACE"
	HEADER_NAME_TRACEID       = "X-IDR-TraceId"
	HEADER_NAME_TS            = "X-IDR-Ts"
	HEADER_NAME_AUTHORIZATION = "Authorization"
	HEADER_NAME_AUTH_USER_ID  = "__auth_user_id"
	HEADER_NAME_AUTH_CLEAR    = "__auth_clear"
	// 正则
	UserNameRegexp = `^[a-zA-Z][a-zA-Z0-9_-]{3,15}$`
	PasswordRegexp = `^[a-fA-F0-9]{32,64}$`
	EmailRegexp    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	//PhoneRegexp    = `^(13[0-9]|14[5|7]|15[0|1|2|3|4|5|6|7|8|9]|18[0|1|2|3|5|6|7|8|9])\d{8}$`
	MobileRegexp = `^(1)\d{10}$`
	PhoneRegexp  = "(^((1)\\d{10})$|^((\\d{7,8})|(\\d{4}|\\d{3})-(\\d{7,8})|(\\d{4}|\\d{3})-(\\d{7,8})-(\\d{4}|\\d{3}|\\d{2}|\\d{1})|(\\d{7,8})-(\\d{4}|\\d{3}|\\d{2}|\\d{1}))$)"
)
