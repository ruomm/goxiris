package xvalidator

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/ruomm/goxframework/gox/corex"
	"github.com/ruomm/goxframework/gox/refx"
	"github.com/ruomm/goxiris/xiris/xresponse"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	xRequest_Parse_Param     = "xreq_param"
	xRequest_Parse_Query     = "xreq_query"
	xRequest_Parse_Header    = "xreq_header"
	xRequest_Option_Response = "resp"
	// 正则
	UserNameRegexp = `^[a-zA-Z][a-zA-Z0-9_-]{3,15}$`
	PasswordRegexp = `^[a-fA-F0-9]{32,64}$`
	EmailRegexp    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	//PhoneRegexp    = `^(13[0-9]|14[5|7]|15[0|1|2|3|4|5|6|7|8|9]|18[0|1|2|3|5|6|7|8|9])\d{8}$`
	MobileRegexp = `^(1)\d{10}$`
	PhoneRegexp  = "(^((1)\\d{10})$|^((\\d{7,8})|(\\d{4}|\\d{3})-(\\d{7,8})|(\\d{4}|\\d{3})-(\\d{7,8})-(\\d{4}|\\d{3}|\\d{2}|\\d{1})|(\\d{7,8})-(\\d{4}|\\d{3}|\\d{2}|\\d{1}))$)"
)

var (
	Validator *validator.Validate
)

func XValidatorInit() error {
	Validator = validator.New()
	err := Validator.RegisterValidation("xcompanyid", xValid_Register_CompanyId)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xlimitstr", xValid_Register_LimitStr)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xusername", xValid_Register_UserName)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xpassword", xValid_Register_Password)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xphone", xValid_Register_Phone)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xmobile", xValid_Register_Mobile)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xtafter", xValid_DateTime_AfterCurrent)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xtbefore", xValid_DateTime_BeforeCurrent)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xqueryday", xValid_Register_QueryDay)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xquerymonth", xValid_Register_QueryMonth)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xquerydate", xValid_Register_QueryDate)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xthhmmss", xValid_Register_xthhmmss)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xthhmm", xValid_Register_xthhmm)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xmutilof", xValid_Register_xmutilof)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xnorepeat", xValid_Register_xnorepeat)
	if err != nil {
		return err
	}
	return nil
}

/*
*
通过xvalid_error注入需要自定义显示的错误信息
*/
func XValidator(u interface{}) (error, *[]xresponse.ParamError) {
	err := Validator.Struct(u)
	return xValidatorProcessErr(u, err)
}

/*
*
通过xvalid_error注入需要自定义显示的错误信息
*/
func xValidatorProcessErr(u interface{}, err error) (error, *[]xresponse.ParamError) {
	if err == nil {
		return nil, nil
	}
	invalid, ok := err.(*validator.InvalidValidationError)
	if ok {
		return errors.New("参数校验错误:" + invalid.Error()), nil
	}
	var paramErrors []xresponse.ParamError
	validationErrs := err.(validator.ValidationErrors)
	for _, validationErr := range validationErrs {
		fieldName := validationErr.Field() //获取是哪个字段不符合格式
		typeOf := reflect.TypeOf(u)
		// 如果是指针，获取其属性
		if typeOf.Kind() == reflect.Ptr {
			typeOf = typeOf.Elem()
		}
		field, ok := typeOf.FieldByName(fieldName) //通过反射获取filed
		errorInfo := ""
		errorKey := ""
		if ok {
			errorInfo = field.Tag.Get("xvalid_error") // 获取field对应的reg_error_info tag值
			errorKeyByTag := xPraseJsonTagName(&field)
			if errorKeyByTag == "" {
				errorKeyByTag = xPraseRefxTagName(&field, xRequest_Parse_Param)
			}
			if errorKeyByTag == "" {
				errorKeyByTag = xPraseRefxTagName(&field, xRequest_Parse_Query)
			}
			if errorKeyByTag == "" {
				errorKeyByTag = xPraseRefxTagName(&field, xRequest_Parse_Header)
			}
			if errorKeyByTag == "" {
				errorKeyByTag = fieldName
			}
			errorKey = errorKeyByTag
		}

		if errorInfo == "" {
			errorMsg := validationErr.Error()
			// Key: 'ConfigSpecCreateReq.DataList[1].SpecId' Error:Field validation for 'SpecId' failed on the 'min' tag
			keyTag := "Key:"
			errTag := "Error:"
			keyIndex := strings.Index(errorMsg, keyTag)
			errIndex := strings.Index(errorMsg, errTag)
			if keyIndex >= 0 && errIndex > keyIndex {
				keyStr := strings.TrimSpace(errorMsg[len(keyTag):errIndex])
				errStr := strings.TrimSpace(errorMsg[errIndex+len(errTag):])
				if strings.HasPrefix(keyStr, "'") {
					keyStr = keyStr[1:]
				}
				if strings.HasSuffix(keyStr, "'") {
					keyStr = keyStr[0 : len(keyStr)-1]
				}
				keyStr = transFieldKey(keyStr)
				if len(keyStr) > 0 {
					errorKey = keyStr
					errorInfo = errStr
				} else {
					errorKey = fieldName
					errorInfo = errorMsg
				}
			} else {
				errorKey = fieldName
				errorInfo = errorMsg
			}
		}
		paramErrors = append(paramErrors, xresponse.ParamError{Field: errorKey, Message: errorInfo})
	}
	return errors.New("参数校验错误"), &paramErrors
}

func xPraseJsonTagName(field *reflect.StructField) string {
	refxTagInfo := field.Tag.Get("json")
	refxTagName, _ := corex.ParseTagToNameOption(refxTagInfo)
	if corex.TagIsValid(refxTagName) && refxTagName != "-" {
		return refxTagName
	} else {
		return ""
	}
}

func xPraseRefxTagName(field *reflect.StructField, refxTagKey string) string {
	refxTagInfo := field.Tag.Get(refxTagKey)
	refxTagName, _ := corex.ParseTagToNameOptionFenHao(refxTagInfo)
	if corex.TagIsValid(refxTagName) && refxTagName != "-" {
		return refxTagName
	} else {
		return ""
	}
}

// 驼峰转下划线工具
//func toSnakeCase(str string) string {
//	str = xvalid_matchNonAlphaNumeric.ReplaceAllString(str, "_")     //非常规字符转化为 _
//	snake := xvalid_matchFirstCap.ReplaceAllString(str, "${1}_${2}") //拆分出连续大写
//	snake = xvalid_matchAllCap.ReplaceAllString(snake, "${1}_${2}")  //拆分单词
//	return strings.ToLower(snake)                                    //全部转小写
//}

func transFieldKey(keyStr string) string {
	if len(keyStr) <= 1 {
		return keyStr
	}
	keys := strings.Split(keyStr, ".")
	if len(keys) <= 1 {
		return strings.ToLower(keyStr[0:1]) + keyStr[1:]
	} else {
		var build strings.Builder
		for i, tmpKey := range keys {
			if i <= 0 {
				continue
			}
			if i > 1 {
				build.WriteString(".")
			}
			tmpLen := len(tmpKey)
			if tmpLen <= 0 {
				continue
			} else if tmpLen == 1 {
				build.WriteString(tmpKey)
			} else {
				build.WriteString(strings.ToLower(tmpKey[0:1]) + tmpKey[1:])
			}
		}
		return build.String()
	}
}

func xValid_Register_CompanyId(fl validator.FieldLevel) bool {
	verificationStr := `^[a-z0-9\-]*$`
	validResult, _ := XValid_Register_Regex(fl, verificationStr)
	return validResult
}

// 必须是用户名
func xValid_Register_UserName(fl validator.FieldLevel) bool {
	verificationStr := UserNameRegexp
	validResult, _ := XValid_Register_Regex(fl, verificationStr)
	return validResult
}

// 必须是密码
func xValid_Register_Password(fl validator.FieldLevel) bool {
	verificationStr := PasswordRegexp
	validResult, _ := XValid_Register_Regex(fl, verificationStr)
	return validResult
}

// 必须手机号码
func xValid_Register_Mobile(fl validator.FieldLevel) bool {
	verificationStr := MobileRegexp
	validResult, _ := XValid_Register_Regex(fl, verificationStr)
	return validResult
}

// 必须电话号码
func xValid_Register_Phone(fl validator.FieldLevel) bool {
	verificationStr := PhoneRegexp
	validResult, _ := XValid_Register_Regex(fl, verificationStr)
	return validResult
}

// 不能存在 单引号、双引号、update、delete 等关键词
func xValid_Register_LimitStr(fl validator.FieldLevel) bool {
	verificationStr := `(?:")|(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	validResult, _ := XValid_Register_Regex_Reverse(fl, verificationStr)
	return validResult
}

// 验证Slice不可以重复，基础数据类型直接验证，结构体可以指定验证指定的field字段。
func xValid_Register_xnorepeat(fl validator.FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		flstring := field.String()
		sliceString := corex.StringToSlice(flstring, ",", true)
		sliceDuplicates := corex.SliceDuplicates(sliceString)
		if sliceDuplicates {
			return false
		} else {
			return true
		}
	default:
		flparam := fl.Param()
		flSlice := field.Interface()
		sliceDuplicates := corex.SliceDuplicatesByKey(flSlice, flparam)
		if sliceDuplicates {
			return false
		} else {
			return true
		}
	}
}

// 验证字符串复选，可以多选，但不可重复
func xValid_Register_xmutilof(fl validator.FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		flparam := fl.Param()
		flstring := field.String()
		if len(flparam) <= 0 || len(flstring) <= 0 {
			return false
		}
		sliceParam := corex.StringToSlice(flparam, " ", false)
		sliceString := corex.StringToSlice(flstring, ",", true)
		sliceDuplicates := corex.SliceDuplicates(sliceString)
		if sliceDuplicates {
			return false
		}
		for _, str := range sliceString {
			if !corex.SliceContains(sliceParam, str) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// 验证时分秒mm-hh-ss格式
func xValid_Register_xthhmmss(fl validator.FieldLevel) bool {
	validResult, _ := XValid_Register_Regex(fl, "^([0-1][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9])$")
	return validResult
}

// 验证时分mm-hh格式
func xValid_Register_xthhmm(fl validator.FieldLevel) bool {
	validResult, _ := XValid_Register_Regex(fl, "^([0-1][0-9]|2[0-3]):([0-5][0-9])$")
	return validResult
}

// 验证查询月份或日期
func xValid_Register_QueryDate(fl validator.FieldLevel) bool {
	if xValid_Register_QueryMonth(fl) || xValid_Register_QueryDay(fl) {
		return true
	} else {
		return false
	}
}

// 验证查询日期
func xValid_Register_QueryDay(fl validator.FieldLevel) bool {
	validResult, flStr := XValid_Register_Regex(fl, "^\\d{4}-\\d{2}-\\d{2}$")
	if !validResult {
		return validResult
	}
	timeArr := strings.Split(flStr, "-")
	year := corex.StrToInt64(timeArr[0])
	month := corex.StrToInt64(timeArr[1])
	day := corex.StrToInt64(timeArr[2])
	if year < 0 || year >= 3000 {
		return false
	}
	if month < 1 || month > 12 {
		return false
	}
	dayCountByMonth := corex.GetDayCountByMonth(int(year), int(month))
	if day < 1 || day > int64(dayCountByMonth) {
		return false
	}
	return validResult
}

// 验证查询月份
func xValid_Register_QueryMonth(fl validator.FieldLevel) bool {
	validResult, flStr := XValid_Register_Regex(fl, "^\\d{4}-\\d{2}$")
	if !validResult {
		return validResult
	}
	timeArr := strings.Split(flStr, "-")
	year := corex.StrToInt64(timeArr[0])
	month := corex.StrToInt64(timeArr[1])
	if year < 0 || year >= 3000 {
		return false
	}
	if month < 1 || month > 12 {
		return false
	}
	return validResult
}

func XValid_Register_Regex(fl validator.FieldLevel, verificationStr string) (bool, string) {
	if verificationStr == "" {
		return false, ""
	}
	field := fl.Field()
	vi := refx.ParseToString(field.Interface(), "tf:2006-01-02 15:04:05")
	if vi == nil {
		return false, ""
	}
	viStr := vi.(string)
	re, err := regexp.Compile(verificationStr)
	if err != nil {
		log.Println(err.Error())
		return false, ""
	}
	return re.MatchString(viStr), viStr
}

func XValid_Register_Regex_Reverse(fl validator.FieldLevel, verificationStr string) (bool, string) {
	if verificationStr == "" {
		return false, ""
	}
	field := fl.Field()
	vi := refx.ParseToString(field.Interface(), "tf:2006-01-02 15:04:05")
	if vi == nil {
		return false, ""
	}
	viStr := vi.(string)
	re, err := regexp.Compile(verificationStr)
	if err != nil {
		log.Println(err.Error())
		return false, viStr
	}
	return !re.MatchString(viStr), viStr
}

func xValid_DateTime_BeforeCurrent(fl validator.FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		result, _ := corex.TimeBeforceCurrent(corex.TIME_PATTERN_STANDARD, field.String())
		return result
	case reflect.Int64:
		return time.UnixMilli(field.Int()).Before(time.Now())
	case reflect.Uint64:
		return time.UnixMilli(field.Int()).Before(time.Now())
	case reflect.Int:
		return time.UnixMilli(field.Int()).Before(time.Now())
	case reflect.Uint:
		return time.UnixMilli(field.Int()).Before(time.Now())
	default:
		return false
	}
}

func xValid_DateTime_AfterCurrent(fl validator.FieldLevel) bool {
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		result, _ := corex.TimeAfterCurrent(corex.TIME_PATTERN_STANDARD, field.String())
		return result
	case reflect.Int64:
		return time.UnixMilli(field.Int()).After(time.Now())
	case reflect.Uint64:
		return time.UnixMilli(field.Int()).After(time.Now())
	case reflect.Int:
		return time.UnixMilli(field.Int()).After(time.Now())
	case reflect.Uint:
		return time.UnixMilli(field.Int()).After(time.Now())
	default:
		return false
	}
}
