package errors

import (
	"log"
	"strconv"
	"testing"

	microerr "github.com/micro/go-micro/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewf(t *testing.T) {
	_, err := strconv.Atoi("asdf")
	log.Printf("original error: %+v", err)

	err1 := Params(Wrap(err, "测试 1")).SetParam("cc", "dd")
	assert.Equal(t, err1.(*Error).RawMsg, "测试 1")
	assert.Equal(t, err1.(*Error).Detail, `测试 1: strconv.Atoi: parsing "asdf": invalid syntax`)
	assert.Equal(t, err1.(*Error).Status, "测试 1[cc=dd]")

	err2 := Params(Wrap(err1, "测试 2")).SetParam("aa", "bb").SetParam("ss", "dd")
	assert.Equal(t, err2.(*Error).RawMsg, "测试 1")
	assert.Equal(t, err2.(*Error).Detail, `测试 2: 测试 1: strconv.Atoi: parsing "asdf": invalid syntax`)
	log.Printf("print for Status: %v", err2.(*Error).Status) // 测试 1[cc=dd,aa=bb,ss=dd]
	log.Printf("print for v: %v", err2)
	log.Printf("print for +v: %+v", err2) // 递归输出 error chain
	log.Printf("print for #v: %#v", err2)

	err3 := Wrapc(err2, 200000, "测试 3")
	log.Printf("print for wrapc: %+v", err3)
	log.Println("detail print for wrapc:", GetDetailError(err3))

	// micro error
	err4 := NewFromMicroErr(err3) // 没有透传价值，当成普通错误处理。
	assert.EqualValues(t, []any{err4.(*Error).Code, err4.(*Error).external}, []any{int32(UNKNOWN_EXCEPTION), false})

	merr := &microerr.Error{
		Id:     "micro error id",
		Code:   1200,
		Detail: "micro error detail",
		Status: "micro error status",
	}
	err5 := NewFromMicroErr(merr) // 透传错误
	assert.EqualValues(t, []any{err5.(*Error).Code, err5.(*Error).external}, []any{merr.Code, true})
	log.Printf("print for +v: %+v", err5)
}

func TestNew(t *testing.T) {
	err := New(1000)
	log.Printf("print for New v: %v", err)
	log.Printf("print for New +v: %+v", err)
	log.Printf("print for New #v: %#v", err)

	err = New(1000, "get error")
	log.Printf("print for NewMsg v: %v", err)
	log.Printf("print for NewMsg +v: %+v", err)
	log.Printf("print for NewMsg #v: %#v", err)

	// test for origin code
	err = New(1000)
	err = Wrapc(err, 1001)
	err = Wrapc(err, 1002, "get error")
	err = Wrapc(err, 1003)
	log.Printf("print for Wrapc +v: %+v", err)

	assert.EqualValues(t, Params(err).GetOriginClearingError().GetErrorCode(), int32(1000)) // 原始错误码
	has, _ := Params(err).SearchClearingError(2000)                                         // error chain 中不存在 code = 2000
	assert.EqualValues(t, has, false)
	has, _ = Params(err).SearchClearingError(1001) // error chain 中存在 code = 1000
	assert.EqualValues(t, has, true)
}
