package frame

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response http response data
type Response struct {
	Code    string      `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Time    time.Time   `json:"time,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

// PageResults http response list data
type PageResults struct {
	Total    int         `json:"total,omitempty"`
	Page     int         `json:"page,omitempty"`
	PageSize int         `json:"page_size,omitempty"`
	Results  interface{} `json:"results,omitempty"`
}

// Success http response ok
// default json
func (ctx *Context) Success(data interface{}) {
	resp := &Response{
		Code:    successCode,
		Message: successMsg,
		Data:    data,
		Time:    time.Now(),
		TraceID: ctx.GetTraceID(),
	}
	ctx.Gtx.JSON(200, resp)
}

// ErrorMsg frame err msg
type ErrorMsg interface {
	// error  code
	GetCode() string
	// real error message, only log
	GetReal() string
	// user reply
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
		TraceID: ctx.GetTraceID(),
	}
	ctx.printRealMsgLog(errMsg.GetReal())
	ctx.Gtx.JSON(http.StatusOK, resp)
}

// HTTPError http response error msg and setting http code
// default json
func (ctx *Context) HTTPError(httpCode int, errMsg ErrorMsg) {
	resp := &Response{
		Code:    errMsg.GetCode(),
		Message: errMsg.GetReply(),
		Data:    nil,
		Time:    time.Now(),
		TraceID: ctx.GetTraceID(),
	}
	ctx.printRealMsgLog(errMsg.GetReal())
	ctx.Gtx.JSON(httpCode, resp)
}

// HTTPError2 http error response
func (ctx *Context) HTTPError2(httpCode int, bussCode, userReply string, realMsg error) {
	resp := &Response{
		Code:    bussCode,
		Message: userReply,
		Data:    nil,
		Time:    time.Now(),
		TraceID: ctx.GetTraceID(),
	}
	ctx.printRealMsgLog(realMsg.Error())
	ctx.Gtx.JSON(httpCode, resp)
}

func (ctx *Context) printRealMsgLog(realMsg string) {
	msg := realMsgs(realMsg).String()
	if len(msg) > 0 {
		ctx.Errorln(msg)
	}
}

// HTTPListSuccess if pageData nil or pageData.Results id empty,auto set []
// http response data or date.results is slice or array
// default json
func (ctx *Context) HTTPListSuccess(pageData *PageResults) {
	emptyPage(pageData)
	resp := &Response{
		Code:    successCode,
		Message: successMsg,
		Data:    pageData,
		Time:    time.Now(),
		TraceID: ctx.GetTraceID(),
	}
	ctx.Gtx.JSON(http.StatusOK, resp)
}

// HTTPListError if pageData nil or pageData.Results id empty,auto set []
// http response data or date.results is slice or array
// default json
func (ctx *Context) HTTPListError(errMsg ErrorMsg) {
	resp := &Response{
		Code:    errMsg.GetCode(),
		Message: errMsg.GetReply(),
		Data:    defaultEmptyPage,
		Time:    time.Now(),
		TraceID: ctx.GetTraceID(),
	}
	ctx.printRealMsgLog(errMsg.GetReal())
	ctx.Gtx.JSON(http.StatusOK, resp)
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

// HTTPListError2 if pageData nil or pageData.Results id empty,auto set []
// http response data or date.results is slice or array
// default json
func (ctx *Context) HTTPListError2(httpCode int, errMsg ErrorMsg) {
	resp := &Response{
		Code:    errMsg.GetCode(),
		Message: errMsg.GetReply(),
		Data:    defaultEmptyPage,
		Time:    time.Now(),
		TraceID: ctx.GetTraceID(),
	}
	ctx.printRealMsgLog(errMsg.GetReal())
	ctx.Gtx.JSON(httpCode, resp)
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
