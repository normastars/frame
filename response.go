package frame

import (
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

// ErrorMsg represents a business error with a user-facing message and internal detail.
type ErrorMsg interface {
	GetCode() string
	GetReal() string
	GetReply() string
}

// buildResponse constructs a Response with common fields filled.
func (ctx *Context) buildResponse(code, msg string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: msg,
		Data:    data,
		Time:    time.Now(),
		TraceID: ctx.GetTraceID(),
	}
}

// logRealError logs the internal error detail with a structured field for diagnostics.
func (ctx *Context) logRealError(realMsg string) {
	if realMsg == "" {
		return
	}
	ctx.WithField("real_reason", realMsg).Error("")
}

// Success returns a 200 response with the given data.
func (ctx *Context) Success(data interface{}) {
	ctx.Gtx.JSON(http.StatusOK, ctx.buildResponse(successCode, successMsg, data))
}

// Error returns a 200 response with a business error (non-HTTP failure).
func (ctx *Context) Error(errMsg ErrorMsg) {
	ctx.logRealError(errMsg.GetReal())
	ctx.Gtx.JSON(http.StatusOK, ctx.buildResponse(errMsg.GetCode(), errMsg.GetReply(), nil))
}

// HTTPError returns an HTTP error response with the given status code.
func (ctx *Context) HTTPError(httpCode int, errMsg ErrorMsg) {
	ctx.logRealError(errMsg.GetReal())
	ctx.Gtx.JSON(httpCode, ctx.buildResponse(errMsg.GetCode(), errMsg.GetReply(), nil))
}

// HTTPError2 returns an HTTP error response with inline error fields (no ErrorMsg interface needed).
func (ctx *Context) HTTPError2(httpCode int, bussCode, userReply string, realMsg error) {
	ctx.logRealError(realMsg.Error())
	ctx.Gtx.JSON(httpCode, ctx.buildResponse(bussCode, userReply, nil))
}

// HTTPListSuccess returns a 200 response with paginated data.
// If pageData is nil or pageData.Results is nil, Results is set to an empty slice.
func (ctx *Context) HTTPListSuccess(pageData *PageResults) {
	ensureResults(pageData)
	ctx.Gtx.JSON(http.StatusOK, ctx.buildResponse(successCode, successMsg, pageData))
}

// HTTPListError returns a 200 response with a business error and an empty page data.
func (ctx *Context) HTTPListError(errMsg ErrorMsg) {
	ctx.logRealError(errMsg.GetReal())
	ctx.Gtx.JSON(http.StatusOK, ctx.buildResponse(errMsg.GetCode(), errMsg.GetReply(), defaultEmptyPage))
}

// HTTPListError2 returns an HTTP error response with the given status code and an empty page data.
func (ctx *Context) HTTPListError2(httpCode int, errMsg ErrorMsg) {
	ctx.logRealError(errMsg.GetReal())
	ctx.Gtx.JSON(httpCode, ctx.buildResponse(errMsg.GetCode(), errMsg.GetReply(), defaultEmptyPage))
}

// ensureResults sets Results to an empty slice if it is nil, so JSON serializes as [] not null.
func ensureResults(pageData *PageResults) {
	if pageData == nil {
		return
	}
	if pageData.Results == nil {
		pageData.Results = make([]interface{}, 0)
	}
}
