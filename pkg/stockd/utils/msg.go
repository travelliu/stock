package utils

import "strings"

const (
	LangZh = "zh"
	LangEn = "en"
)

const (
	SUCCESS = 200
	ERROR   = 500

	ErrInvalidParam    = 40001
	ErrUnauthorized    = 40002
	ErrForbidden       = 40003
	ErrUserNotFound    = 40004
	ErrInvalidPassword = 40005
	ErrUserDisabled    = 40006
	ErrStockNotFound   = 40007
	ErrTokenInvalid    = 40009
	ErrTokenExpired    = 40010
	ErrDuplicateUser   = 40011
	ErrInvalidCode     = 40012
	ErrTaskRun         = 40013
	ErrTaskNoRunReport = 40014
)

var (
	defaultZhMsg = map[int]string{
		SUCCESS:            "成功",
		ERROR:              "系统异常，请联系管理员",
		ErrInvalidParam:    "参数错误",
		ErrUnauthorized:    "未认证",
		ErrForbidden:       "无权限",
		ErrUserNotFound:    "用户不存在",
		ErrInvalidPassword: "密码错误",
		ErrUserDisabled:    "账户已禁用",
		ErrStockNotFound:   "股票不存在",
		ErrTokenInvalid:    "Token无效",
		ErrTokenExpired:    "Token已过期",
		ErrDuplicateUser:   "用户名已存在",
		ErrInvalidCode:     "股票代码错误",
		ErrTaskRun:         "任务正在运行",
		ErrTaskNoRunReport: "任务没有运行报告",
	}
	defaultEnMsg = map[int]string{
		SUCCESS:            "Succeed",
		ERROR:              "The system is abnormal, please contact the administrator",
		ErrInvalidParam:    "Invalid parameter",
		ErrUnauthorized:    "Unauthorized",
		ErrForbidden:       "Forbidden",
		ErrUserNotFound:    "User not found",
		ErrInvalidPassword: "Invalid password",
		ErrUserDisabled:    "Account disabled",
		ErrStockNotFound:   "Stock not found",
		ErrTokenInvalid:    "Invalid token",
		ErrTokenExpired:    "Token expired",
		ErrDuplicateUser:   "Username already exists",
		ErrInvalidCode:     "Invalid stock code",
		ErrTaskRun:         "Task is running",
		ErrTaskNoRunReport: "Task has no run report",
	}

	defaultErrorMsg = map[string]map[int]string{
		LangZh: defaultZhMsg,
		LangEn: defaultEnMsg,
	}
	ErrorMsg = map[string]map[int]string{
		LangZh: defaultZhMsg,
		LangEn: defaultEnMsg,
	}
)

func GetErrMsg(code int, lang string) string     { return getMsg(code, lang, ERROR) }
func GetSuccessMsg(code int, lang string) string { return getMsg(code, lang, SUCCESS) }
func GetMsg(code int, lang string) string        { return getMsg(code, lang, 0) }

func getMsg(code int, lang string, status int) string {
	lang = strings.ToLower(lang)
	if lang == "" {
		lang = LangEn
	}
	msgMap, ok := ErrorMsg[lang]
	if !ok {
		msgMap = defaultErrorMsg[LangEn]
	}
	if m, ok := msgMap[code]; ok {
		return m
	}
	if status != 0 {
		if m, ok := msgMap[status]; ok {
			return m
		}
	}
	return ""
}
