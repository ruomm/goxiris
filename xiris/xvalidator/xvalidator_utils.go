package xvalidator

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ruomm/goxframework/gox/corex"
	"github.com/ruomm/goxframework/gox/refx"
	"github.com/ruomm/goxiris/xiris/xresponse"
	"log"
	"reflect"
	"regexp"
	"strconv"
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

// 初始化validator.Validate
func XValidatorInit() error {
	//
	/**
	xcompanyid：^[a-z0-9\-]*$
	xlimitstr：不能存在 单引号、双引号、update、delete 等关键词
	xusername：^[a-zA-Z][a-zA-Z0-9_-]{3,15}$
	xpassword：^[a-fA-F0-9]{32,64}$
	xphone：(^((1)\d{10})$|^((\d{7,8})|(\d{4}|\d{3})-(\d{7,8})|(\d{4}|\d{3})-(\d{7,8})-(\d{4}|\d{3}|\d{2}|\d{1})|(\d{7,8})-(\d{4}|\d{3}|\d{2}|\d{1}))$)，手机号码、电话号码
	xmobile：`^(1)\d{10}$，手机号码
	xtafter：在当前时间之后
	xtbefore：在当前时间之前
	xqueryday：yyyy-MM-dd格式日期
	xquerymonth：yyyy-MM格式月份
	xquerydate：yyyy-MM-dd格式日期、yyyy-MM格式月份
	xquerytime：yyyy-MM-dd HH:mm:ss格式时间
	xquerydayortime：yyyy-MM-dd格式日期、yyyy-MM-dd HH:mm:ss格式时间
	xthhmmss：HH:mm:ss格式时间
	xthhmm：mm:ss格式时间
	xmutilof：验证字符串复选，可以多选，但不可重复
	xnorepeat：验证Slice不可以重复，基础数据类型直接验证，结构体可以指定验证指定的field字段
	xweburl：验证WEB网址，必须是https或http协议，没有参数协议后面至少1位字符串，有参数则协议后字符串长度必须大于等于参数值
	xfilename：验证是否文件，不能包含文件分隔符、换行符、tab字符，必须包含.字符，不能以.字符开头结束
	*/

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
	err = Validator.RegisterValidation("xquerytime", xValid_Register_QueryTime)
	if err != nil {
		return err
	}

	err = Validator.RegisterValidation("xquerydayortime", xValid_Register_QueryDayOrTime)
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
	err = Validator.RegisterValidation("xweburl", xValid_Register_xweburl)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xfilename", xValid_Register_xfilename)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xfilepath", xValid_Register_xfilepath)
	if err != nil {
		return err
	}
	err = Validator.RegisterValidation("xfiledir", xValid_Register_xfiledir)
	if err != nil {
		return err
	}
	return nil
}

/*
*
通过xvalid_error注入需要自定义显示的错误信息，showFirstError：true时候显示第一条具体错误信息，false时候显示参数校验错误
*/
func XValidator(u interface{}, showFirstError bool) (error, *[]xresponse.ParamError) {
	err := Validator.Struct(u)
	return xValidatorProcessErr(u, err, showFirstError)
}

/*
*
通过xvalid_error注入需要自定义显示的错误信息
*/
func xValidatorProcessErr(u interface{}, err error, showFirstError bool) (error, *[]xresponse.ParamError) {
	if err == nil {
		return nil, nil
	}
	invalid, okG := err.(*validator.InvalidValidationError)
	if okG {
		if showFirstError {
			return errors.New("参数校验器失效:" + invalid.Error()), nil
		} else {
			return errors.New("参数校验错误:" + invalid.Error()), nil
		}

	}
	var paramErrors []xresponse.ParamError
	validationErrs := err.(validator.ValidationErrors)
	firstErrorInfo := ""
	if !showFirstError {
		firstErrorInfo = "参数校验错误"
	}
	for _, validationErr := range validationErrs {
		field, parentFileName, fieldName, parseErr := xParseStructField(u, validationErr)
		errorInfo := ""
		errorKey := ""
		if parseErr == nil {
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
			if len(errorKeyByTag) > 0 {
				errorKey = parentFileName + errorKeyByTag
			}
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
		if showFirstError && len(firstErrorInfo) <= 0 {
			firstErrorInfo = fmt.Sprintf("%s", errorInfo)
		}
		paramErrors = append(paramErrors, xresponse.ParamError{Field: errorKey, Message: errorInfo})
	}
	if len(firstErrorInfo) <= 0 {
		firstErrorInfo = "参数校验错误"
	}
	return errors.New(firstErrorInfo), &paramErrors
}

func xParseStructField(u interface{}, validationErr validator.FieldError) (reflect.StructField, string, string, error) {
	ns := validationErr.Namespace()
	_, fieldPath, _ := strings.Cut(ns, ".")
	fieldNames := strings.Split(fieldPath, ".")
	if len(fieldPath) <= 0 || len(fieldNames) <= 0 {
		typeOf := reflect.TypeOf(u)
		typeKind := typeOf.Kind()
		// 如果是指针，获取其属性
		if typeKind == reflect.Slice || typeKind == reflect.Pointer || typeKind == reflect.Map || typeKind == reflect.Chan || typeKind == reflect.Array {
			typeOf = typeOf.Elem()
		}
		fieldName := validationErr.Field()
		field, ok := typeOf.FieldByName(fieldName) //通过反射获取filed
		if !ok {
			return field, "", fieldName, errors.New(fmt.Sprintf("无法获取%s字段的类型", fieldName))
		} else {
			return field, "", fieldName, nil
		}
	} else {
		typeOf := reflect.TypeOf(u)
		parentFieldBuilder := strings.Builder{}
		subFieldName := fieldNames[len(fieldNames)-1]
		for i := 0; i < len(fieldNames)-1; i++ {
			parentFieldBuilder.WriteString(corex.FirstLetterToLower(fieldNames[i]))
			parentFieldBuilder.WriteString(".")
		}
		for i := 0; i < len(fieldNames); i++ {
			fieldName, _ := xParseRealFieldName(fieldNames[i])
			// 如果是指针，获取其属性
			typeKind := typeOf.Kind()
			if typeKind == reflect.Slice || typeKind == reflect.Pointer || typeKind == reflect.Map || typeKind == reflect.Chan || typeKind == reflect.Array {
				typeOf = typeOf.Elem()
			}
			field, ok := typeOf.FieldByName(fieldName) //通过反射获取filed
			if !ok {
				return field, parentFieldBuilder.String(), subFieldName, errors.New(fmt.Sprintf("无法获取%s字段的类型", fieldPath))
			}
			typeOf = field.Type
			if i == len(fieldNames)-1 {
				return field, parentFieldBuilder.String(), subFieldName, nil
			}
		}
		return reflect.StructField{}, parentFieldBuilder.String(), subFieldName, errors.New(fmt.Sprintf("无法获取%s字段的类型", fieldPath))
	}

}
func xParseRealFieldName(fieldName string) (string, bool) {
	if strings.HasSuffix(fieldName, "]") {
		lastIndex := strings.LastIndex(fieldName, "[")
		if lastIndex > 0 {
			return fieldName[0:lastIndex], true
		} else {
			return fieldName, false
		}
	} else {
		return fieldName, false
	}
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
		sliceParam := corex.StringToSlice(flparam, " ", false)
		sliceDuplicates := true
		if len(sliceParam) <= 0 {
			sliceDuplicates = corex.SliceDuplicates(flSlice)
		} else {
			sliceDuplicates = corex.SliceDuplicatesByKey(flSlice, sliceParam...)
		}
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

// 验证查询日期
func xValid_Register_QueryDayOrTime(fl validator.FieldLevel) bool {
	flString, _ := XValidParseToString(fl)
	flStrLen := len(flString)
	if flStrLen != 10 && flStrLen != 19 {
		return false
	}
	if flStrLen == 10 {
		flString = flString + " 00:00:00"
	}
	timeParse, err := corex.TimeParseByString(corex.TIME_PATTERN_STANDARD, flString)
	if err != nil {
		return false
	}
	if timeParse == nil {
		return false
	}
	return true
}

// 验证查询日期
func xValid_Register_QueryTime(fl validator.FieldLevel) bool {
	flString, _ := XValidParseToString(fl)
	flStrLen := len(flString)
	if flStrLen != 19 {
		return false
	}
	timeParse, err := corex.TimeParseByString(corex.TIME_PATTERN_STANDARD, flString)
	if err != nil {
		return false
	}
	if timeParse == nil {
		return false
	}
	return true
}

// 验证WEB网址，必须是https或http协议，没有参数协议后面至少1位字符串，有参数则协议后字符串长度必须大于等于参数值
func xValid_Register_xweburl(fl validator.FieldLevel) bool {
	flString, _ := XValidParseToString(fl)
	if len(flString) <= 0 {
		return false
	}
	flStringLower := strings.ToLower(flString)
	lenFlString := len(flStringLower)
	flParamInt := 1
	flParam := fl.Param()
	if len(flParam) > 0 {
		i, _ := strconv.ParseInt(flParam, 10, 64)
		flParamInt = int(i)
	}
	if strings.HasPrefix(flStringLower, "http://") && lenFlString >= 7+flParamInt {
		return true
	} else if strings.HasPrefix(flStringLower, "https://") && lenFlString >= 8+flParamInt {
		return true
	} else {
		return false
	}
}

// 验证WEB网址，必须是https或http协议，没有参数协议后面至少1位字符串，由参数则协议后字符串长度必须大于等于参数值
func old_xValid_Register_xfilename(fl validator.FieldLevel) bool {
	validResult, flStr := XValid_Register_Regex(fl, "^[a-zA-Z0-9-_\\.]{1,255}$")
	if !validResult {
		return validResult
	}
	if strings.HasSuffix(flStr, ".") {
		return false
	}
	if strings.HasPrefix(flStr, ".") {
		return false
	}
	flParam := fl.Param()
	if len(flParam) > 0 {
		flParam = strings.ToLower(flParam)
	}
	if flParam == "false" {
		return true
	}
	fileNameWithoutExtension := corex.GetFileNameWithoutExtension(flStr)
	fileExtension := corex.GetFileExtension(flStr)
	if len(fileNameWithoutExtension) <= 0 || len(fileExtension) <= 0 {
		return false
	} else {
		return true
	}
}

// 验证是否文件，不能包含文件分隔符、换行符、tab字符，必须包含.字符
func xValid_Register_xfilename(fl validator.FieldLevel) bool {
	fileName, _ := XValidParseToString(fl)
	if len(fileName) <= 0 {
		return false
	}
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") || strings.Contains(fileName, "\r") || strings.Contains(fileName, "\n") || strings.Contains(fileName, "\t") {
		return false
	}
	fileNameWithoutExtension := corex.GetFileNameWithoutExtension(fileName)
	fileExtension := strings.ToLower(corex.GetFileExtension(fileName))
	if len(fileNameWithoutExtension) <= 0 {
		return false
	}
	if len(fileExtension) <= 0 {
		return false
	}
	flParam := fl.Param()
	if len(flParam) <= 0 {
		return true
	}
	sliceParam := corex.StringToSlice(flParam, " ", false)
	if len(sliceParam) <= 0 {
		return true
	}
	flResult := false
	for _, tmpExt := range sliceParam {
		tmpExtLower := strings.ToLower(tmpExt)
		if fileExtension == tmpExtLower || "."+fileExtension == tmpExtLower {
			flResult = true
			break
		}
	}
	return flResult
}

// 验证是否文件，不能包含文件分隔符、换行符、tab字符，必须包含.字符
func xValid_Register_xfilepath(fl validator.FieldLevel) bool {
	filePath, _ := XValidParseToString(fl)
	if len(filePath) <= 0 {
		return false
	}
	if strings.Contains(filePath, "\r") || strings.Contains(filePath, "\n") || strings.Contains(filePath, "\t") {
		return false
	}
	fileName := corex.GetFileName(filePath)
	fileNameWithoutExtension := corex.GetFileNameWithoutExtension(fileName)
	fileExtension := strings.ToLower(corex.GetFileExtension(fileName))
	if len(fileNameWithoutExtension) <= 0 {
		return false
	}
	if len(fileExtension) <= 0 {
		return false
	}
	flParam := fl.Param()
	if len(flParam) <= 0 {
		return true
	}
	sliceParam := corex.StringToSlice(flParam, " ", false)
	if len(sliceParam) <= 0 {
		return true
	}
	flResult := false
	for _, tmpExt := range sliceParam {
		tmpExtLower := strings.ToLower(tmpExt)
		if fileExtension == tmpExtLower || "."+fileExtension == tmpExtLower {
			flResult = true
			break
		}
	}
	return flResult
}

// 验证是否文件夹，不能包含文件分隔符、换行符、tab字符，最后的子文件夹中不能含有.符号
func xValid_Register_xfiledir(fl validator.FieldLevel) bool {
	filePath, _ := XValidParseToString(fl)
	if len(filePath) <= 0 {
		return false
	}
	if strings.Contains(filePath, "\r") || strings.Contains(filePath, "\n") || strings.Contains(filePath, "\t") {
		return false
	}
	fileName := corex.GetFileName(filePath)
	if len(fileName) <= 0 {
		return true
	}
	if strings.Contains(fileName, ".") {
		return false
	} else {
		return true
	}
	//if strings.HasSuffix(fileName, ".") {
	//	return false
	//}
	//fileNameWithoutExtension := corex.GetFileNameWithoutExtension(fileName)
	//fileExtension := strings.ToLower(corex.GetFileExtension(fileName))
	//if len(fileNameWithoutExtension) <= 0 {
	//	return true
	//}
	//if len(fileExtension) <= 0 {
	//	return true
	//} else {
	//	return false
	//}
}

// 解析为字符串
func XValidParseToString(fl validator.FieldLevel) (string, error) {

	field := fl.Field()
	vi := refx.ParseToString(field.Interface(), "tf:2006-01-02 15:04:05")
	if vi == nil {
		return "", errors.New("validator field can not parse to string")
	}
	viStr := vi.(string)
	return viStr, nil
}

// 正则验证器匹配验证
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

// 正则验证器不匹配验证
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
