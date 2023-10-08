package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/micro/go-micro/config"
	microerr "github.com/micro/go-micro/errors"
)

type ClearingError interface {
	GetId() string
	GetErrorCode() int32
	GetOriginClearingError() ClearingError
	SearchClearingError(code int32) (bool, ClearingError)
	GetRawMsg() string
	SetRawMsg(string)
	SetDetail(string)
	GetDetail() string
	SetParam(string, interface{}) ClearingError
	GetParams() map[string]string
	Error() string
	To() error
	GetStatus() string
}

const (
	UNKNOWN_EXCEPTION = 99999
)

// 打印详细日志信息
func GetDetailError(err error, level ...int) string {
	type causer interface {
		Cause() error
	}

	if err == nil {
		return ""
	}

	l := 0
	if len(level) > 0 {
		l = level[0]
	}

	str := fmt.Sprintf("|-%s%#v", strings.Repeat("-", l*2), err)
	cau, ok := err.(causer)
	if !ok {
		return str
	}
	return str + "\n" + GetDetailError(cau.Cause(), l+1)
}

// 将 error 接口对象包装成 ClearingError 接口对象
func Params(err error) ClearingError {
	cerr, ok := err.(ClearingError)
	if !ok {
		err = Wrap(err, "Unknown exception")
		cerr, _ = err.(ClearingError)
	}
	return cerr
}

func IsClearingError(err error) bool {
	_, ok := err.(ClearingError)
	return ok
}

type Error struct {
	Id         string
	Code       int32
	Detail     string            // 后台链路信息
	Status     string            // 返回给前端展示的 message
	RawMsg     string            // 保留最原始的错误信息
	Params     map[string]string // 保存额外的错误参数
	causeError error
	external   bool // 标记 code 是否来自于外部系统透传
}

func (e *Error) GetErrorCode() int32 {
	return e.Code
}
func (e *Error) GetStatus() string {
	return e.Status
}

// 获取最底层的 ClearingError
func (e *Error) GetOriginClearingError() ClearingError {
	if cerr, ok := e.causeError.(ClearingError); ok {
		return cerr.GetOriginClearingError()
	}
	return e
}

// 根据 code 搜索 ClearingError
func (e *Error) SearchClearingError(code int32) (bool, ClearingError) {
	if e.Code == code {
		return true, e
	}
	if cerr, ok := e.causeError.(ClearingError); ok {
		return cerr.SearchClearingError(code)
	}
	return false, nil
}

func (e *Error) GetRawMsg() string {
	return e.RawMsg
}

func (e *Error) SetRawMsg(rawMsg string) {
	e.RawMsg = rawMsg
	e.updateStatus()
}

func (e *Error) SetDetail(detail string) {
	e.Detail = detail
}

func (e *Error) GetDetail() string {
	return e.Detail
}

// ClearingError 接口对象转成 error 接口对象
func (e *Error) To() error {
	return e
}

// 内部方法，拼接参数信息为字符串
func (e *Error) getParamString() string {
	if len(e.Params) == 0 {
		return ""
	}
	res := make([]string, len(e.Params))
	for key, value := range e.Params {
		res = append(res, fmt.Sprintf("%s=%s", key, value))
	}
	return fmt.Sprintf("[%s]", strings.TrimLeft(strings.Join(res, ","), ","))
}

// 内部方法，更新 status 字段，RawMsg + 参数信息
func (e *Error) updateStatus() {
	e.Status = e.RawMsg + e.getParamString()
}

func (e *Error) SetParam(key string, value interface{}) ClearingError {
	e.Params[key] = fmt.Sprintf("%v", value)
	e.updateStatus()
	return e
}

func (e *Error) GetParams() map[string]string {
	return e.Params
}

func (e *Error) GetId() string {
	return e.Id
}

func (e *Error) Error() string {
	//if e.causeError == nil {
	//	return e.Detail
	//}
	//return e.Detail + ": " + e.causeError.Error()
	// 请勿修改，网关会识别返回值，不是 json 无法解析，会当做未知异常
	b, _ := json.Marshal(e)
	return string(b)
}

// implment fmt.Formatter
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') { // 递归输出 error chain
			b, _ := json.Marshal(e)
			io.WriteString(s, string(b))
			io.WriteString(s, "\n") // 换行
			if e.causeError != nil {
				fmt.Fprintf(s, "%+v", e.causeError)
			}
			return
		}
		if s.Flag('#') { // json marshal 输出
			b, _ := json.Marshal(e)
			io.WriteString(s, string(b))
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, e.Error())
	}
}

func (e *Error) Cause() error {
	return e.causeError
}

// New create new micro errors
func New(code int32, msg ...string) error {
	m := ""
	if len(msg) > 0 {
		m = strings.Join(msg, ", ")
	} else {
		m = fmt.Sprintf("err_code=%d", code) // 默认文案
	}
	e := &Error{
		Id:     config.Get("micro", "server", "name").String("unknown"),
		Code:   code,
		Detail: m,
		Params: map[string]string{},
		RawMsg: m,
	}
	e.updateStatus()
	return e
}

// Errorf create new micro errors
func Errorf(code int32, format string, args ...interface{}) error {
	return New(code, fmt.Sprintf(format, args...))
}

// NewFromMicroErr .
func NewFromMicroErr(err error) error {
	merr, ok := err.(*microerr.Error)
	// status 为空的 error 没有透传价值，当成普通错误处理。
	if !ok || merr.Status == "" {
		return New(int32(UNKNOWN_EXCEPTION), err.Error())
	}
	e := &Error{
		Id:       merr.Id,
		Code:     merr.Code,
		Detail:   merr.Detail,
		Params:   map[string]string{},
		RawMsg:   merr.Status,
		external: true, // 标记 code 来自于外部系统透传
	}
	e.updateStatus()
	return e
}

// Wrapc If err is nil, Wrapf returns nil.
func Wrapc(err error, code int32, msg ...string) error {
	if err == nil {
		return nil
	}
	m := ""
	if len(msg) > 0 {
		m = strings.Join(msg, ", ")
	} else {
		m = fmt.Sprintf("err_code=%d", code) // 默认文案
	}
	return Wrapcf(err, code, m)
}

// Wrapcf If err is nil, Wrapf returns nil.
func Wrapcf(err error, code int32, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	var detail string
	message := fmt.Sprintf(format, args...)
	cerr, ok := err.(ClearingError)
	if !ok {
		detail = message + ": " + err.Error()
	} else {
		detail = message + ": " + cerr.GetDetail()
	}

	e := &Error{
		Id:         config.Get("micro", "server", "name").String("unknown"),
		Code:       code,
		Detail:     detail,
		causeError: err,
		Params:     map[string]string{},
		RawMsg:     message,
	}
	e.updateStatus()
	return e
}

// Wrap If err is nil, Wrapf returns nil.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	return Wrapf(err, msg)
}

// Wrapf If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	cerr, ok := err.(ClearingError)
	if !ok {
		return Wrapcf(err, int32(UNKNOWN_EXCEPTION), format, args...)
	} else {
		message := fmt.Sprintf(format, args...) + ": " + cerr.GetDetail()
		// 复用 code, params, rawMsg 等信息，构造一个新的 ClearingError 返回，不修改原有的 err
		e := &Error{
			Id:         cerr.GetId(),
			Code:       cerr.GetErrorCode(),
			Detail:     message,
			causeError: err,
			Params:     cerr.GetParams(),
			RawMsg:     cerr.GetRawMsg(),
		}
		e.updateStatus()
		return e.To()
	}
}

func Is(err, target error) bool { return errors.Is(err, target) }

func GetShortMessage(err error) string {
	message := err.Error()
	clearingError, ok := err.(ClearingError)
	if ok {
		message = clearingError.GetStatus()
	}
	if len(message) > 1900 {
		message = message[:1900] // 保留 1900 字符
	}
	return message
}

// 获取前台展示的错误信息
func GetFrontErrorMessageByLang(err error, lang string) string {
	_, ok := err.(ClearingError)
	if !ok {
		err = Wrap(err, "unknown exception")
	}

	clearingError, ok := err.(ClearingError)

	TransferMsg(lang, clearingError, ERROR_MAPS)
	return clearingError.GetRawMsg()
}
