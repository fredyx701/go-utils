package errors

import (
	"context"
	"log"

	"github.com/micro/go-micro/server"
)

// 全局错误码报保存
var ERROR_MAPS = make(map[int32]*ErrorMsg)

type ErrorMsg struct {
	En   string
	ZhCN string
	ZhHK string
}

type Option struct {
	ErrorMaps   map[int32]*ErrorMsg
	LoggerError bool
}

type OptionFunc func(*Option)

func WithErrorMaps(errorMaps map[int32]*ErrorMsg) OptionFunc {
	return func(option *Option) {
		option.ErrorMaps = errorMaps
	}
}

func WithLoggerError(flag bool) OptionFunc {
	return func(option *Option) {
		option.LoggerError = flag
	}
}

// 方法白名单
var accountChannelMapWhite = map[string]struct{}{
	"Debug.Health": {},
}

func NewErrorWrapper(opts ...OptionFunc) server.HandlerWrapper {
	var options Option
	options.LoggerError = true
	for _, opt := range opts {
		opt(&options)
	}
	if options.ErrorMaps == nil {
		options.ErrorMaps = make(map[int32]*ErrorMsg)
	}
	options.ErrorMaps[UNKNOWN_EXCEPTION] = &ErrorMsg{
		En:   "Click to provide feedback to Whale, and an engineer will follow up to address your issue.",
		ZhCN: "点击反馈给 Whale，将会有工程师跟进处理你的问题。",
		ZhHK: "點擊反饋給 Whale，將會有工程師跟進處理你的問題。",
	}

	ERROR_MAPS = options.ErrorMaps

	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			_, has := accountChannelMapWhite[req.Method()]
			if has {
				return fn(ctx, req, rsp)
			}

			err := fn(ctx, req, rsp)
			if err != nil {
				_, ok := err.(ClearingError)
				if !ok {
					err = Wrap(err, "Unknown exception")
				}
				//i18n
				cerr, _ := err.(ClearingError)
				lang := "zh-CN"
				TransferMsg(lang, cerr, options.ErrorMaps)
				if options.LoggerError {
					log.Println(GetDetailError(err))
				}
			}

			return err
		}
	}
}

// TransferMsg 多语言转换文案
func TransferMsg(lang string, cerr ClearingError, errorMaps map[int32]*ErrorMsg) {
	// 外部系统透传的 code 直接返回
	if terr, ok := cerr.(*Error); ok && terr.external {
		return
	}

	errorMsg, ok := errorMaps[cerr.GetErrorCode()]
	if ok {
		msg := errorMsg.En
		switch lang {
		case "zh-CN":
			msg = errorMsg.ZhCN
		case "zh-HK":
			msg = errorMsg.ZhHK
		}
		cerr.SetRawMsg(msg)
	}
	// 匹配不上直接返回
}
