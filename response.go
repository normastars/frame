package frame

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response http response data
type Response struct {
	Code      string      `json:"code,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Time      time.Time   `json:"time,omitempty"`
}

// PageResults http response list data
type PageResults struct {
	Total    int           `json:"total,omitempty"`
	Page     int           `json:"page,omitempty"`
	PageSize int           `json:"page_size,omitempty"`
	Results  []interface{} `json:"results,omitempty"`
}

var (
	successMsg       = "ok"
	successCode      = "0"
	defaultEmptyPage = PageResults{
		Results: make([]interface{}, 0),
	}
)

// Success http response 响应成功时使用
// default json
func (ctx *Context) Success(data interface{}) {
	resp := &Response{
		Code:    successCode,
		Message: successMsg,
		Data:    data,
		Time:    time.Now(),
	}
	ctx.JSON(200, resp)
}

// ErrorMsg frame err msg
type ErrorMsg interface {
	// 错误码
	GetCode() string
	// 真实原因, 仅用于打印日志
	GetReal() string
	// 对客户描述,对外返回
	GetReply() string
}

// Error http response error msg
// default json
func (ctx *Context) Error(errMsg ErrorMsg) {
	resp := &Response{
		Code:    errMsg.GetCode(),
		Message: errMsg.GetReply(),
		Data:    nil,
		Time:    time.Now(),
	}
	ctx.printRealMsgLog(errMsg.GetReal())
	ctx.JSON(http.StatusOK, resp)
}

// HTTPError http response error msg and setting http code
// default json
func (ctx *Context) HTTPError(httpCode int, errMsg ErrorMsg) {
	resp := &Response{
		RequestID: ctx.GetTraceID(),
		Code:      errMsg.GetCode(),
		Message:   errMsg.GetReply(),
		Data:      nil,
		Time:      time.Now(),
	}
	ctx.printRealMsgLog(errMsg.GetReal())
	ctx.JSON(httpCode, resp)
}

// HTTPError2 http error response
func (ctx *Context) HTTPError2(httpCode int, bussCode, userReply string, realMsg error) {
	resp := &Response{
		RequestID: ctx.GetTraceID(),
		Code:      bussCode,
		Message:   userReply,
		Data:      nil,
		Time:      time.Now(),
	}
	ctx.printRealMsgLog(realMsg.Error())
	ctx.JSON(httpCode, resp)
}

func (ctx *Context) printRealMsgLog(realMsg string) {
	msg := realMsgs(realMsg).String()
	if len(msg) > 0 {
		ctx.Errorln(msg)
	}
}

// HTTPListSuccess 如果 pageData 是nil 或者 pageData.Results 是空,自动设置为空数组[]
// http response 结果集是数组或切片时使用
// default json
func (ctx *Context) HTTPListSuccess(pageData *PageResults) {
	emptyPage(pageData)
	resp := &Response{
		RequestID: ctx.GetTraceID(),
		Code:      successCode,
		Message:   successMsg,
		Data:      pageData,
		Time:      time.Now(),
	}
	ctx.JSON(http.StatusOK, resp)
}

// HTTPListError 自动将结果集设置为空数组
// http response 结果集是数组或切片时使用
// default json
func (ctx *Context) HTTPListError(errMsg ErrorMsg) {
	resp := &Response{
		Code:    errMsg.GetCode(),
		Message: errMsg.GetReply(),
		Data:    defaultEmptyPage,
		Time:    time.Now(),
	}
	ctx.printRealMsgLog(errMsg.GetReal())
	ctx.JSON(http.StatusOK, resp)
}

func realMsgs(msg string) *realMsg {
	return &realMsg{
		Mode: "real_reason",
		Msg:  msg,
	}
}

type realMsg struct {
	Mode string `json:"mode,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

func (m *realMsg) String() string {
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}

// HTTPListError2 自动将结果集设置为空数组
// http response 结果集是数组或切片时使用
// default json
func (ctx *Context) HTTPListError2(httpCode int, errMsg ErrorMsg) {
	resp := &Response{
		RequestID: ctx.GetTraceID(),
		Code:      errMsg.GetCode(),
		Message:   errMsg.GetReply(),
		Data:      defaultEmptyPage,
		Time:      time.Now(),
	}
	ctx.printRealMsgLog(errMsg.GetReal())
	ctx.JSON(httpCode, resp)
}

func emptyPage(pageData *PageResults) {
	if !(pageData != nil) {
		pageData = &defaultEmptyPage
		return
	}
	if pageData.Results == nil {
		pageData.Results = make([]interface{}, 0)
	}
}
