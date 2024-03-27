package xvalidator

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/ruomm/goxframework/gox/corex"
	"github.com/ruomm/goxiris/xiris/xresponse"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	xRequest_Parse_Param_COMMON = "xreq_param"
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
			jsonInfo := field.Tag.Get("json")
			jsonName, _ := corex.ParseTagToNameOption(jsonInfo)
			if corex.TagIsValid(jsonName) {
				errorKey = jsonName
			} else {
				xreqTag := field.Tag.Get(xRequest_Parse_Param_COMMON)
				xreqKey, _ := corex.ParseTagToNameOption(xreqTag)
				if corex.TagIsValid(xreqKey) {
					errorKey = xreqKey
				} else {
					errorKey = fieldName
				}
			}
		} else {

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
				errStr := strings.TrimSpace(errTag[errIndex+len(errTag):])
				if strings.HasPrefix(keyStr, "'") {
					keyStr = keyStr[1:]
				}
				if strings.HasSuffix(keyStr, "'") {
					keyStr = keyStr[0 : len(keyStr)-1]
				}
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

// 驼峰转下划线工具
//func toSnakeCase(str string) string {
//	str = xvalid_matchNonAlphaNumeric.ReplaceAllString(str, "_")     //非常规字符转化为 _
//	snake := xvalid_matchFirstCap.ReplaceAllString(str, "${1}_${2}") //拆分出连续大写
//	snake = xvalid_matchAllCap.ReplaceAllString(snake, "${1}_${2}")  //拆分单词
//	return strings.ToLower(snake)                                    //全部转小写
//}

func xValid_Register_CompanyId(fl validator.FieldLevel) bool {
	verificationStr := `^[a-z0-9\-]*$`
	return xValid_Register_Regex(fl, verificationStr)
}

// 必须是用户名
func xValid_Register_UserName(fl validator.FieldLevel) bool {
	verificationStr := UserNameRegexp
	return xValid_Register_Regex(fl, verificationStr)
}

// 必须是密码
func xValid_Register_Password(fl validator.FieldLevel) bool {
	verificationStr := PasswordRegexp
	return xValid_Register_Regex(fl, verificationStr)
}

// 必须手机号码
func xValid_Register_Mobile(fl validator.FieldLevel) bool {
	verificationStr := MobileRegexp
	return xValid_Register_Regex(fl, verificationStr)
}

// 必须电话号码
func xValid_Register_Phone(fl validator.FieldLevel) bool {
	verificationStr := PhoneRegexp
	return xValid_Register_Regex(fl, verificationStr)
}

// 不能存在 单引号、双引号、update、delete 等关键词
func xValid_Register_LimitStr(fl validator.FieldLevel) bool {
	verificationStr := `(?:")|(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	return xValid_Register_Regex_Reverse(fl, verificationStr)
}

func xValid_Register_Regex(fl validator.FieldLevel, verificationStr string) bool {
	if verificationStr == "" {
		return false
	}
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		re, err := regexp.Compile(verificationStr)
		if err != nil {
			log.Println(err.Error())
			return false
		}
		return re.MatchString(field.String())
	default:
		return false
	}
}

func xValid_Register_Regex_Reverse(fl validator.FieldLevel, verificationStr string) bool {
	if verificationStr == "" {
		return false
	}
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		re, err := regexp.Compile(verificationStr)
		if err != nil {
			log.Println(err.Error())
			return false
		}
		return !re.MatchString(field.String())
	default:
		return false
	}
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
