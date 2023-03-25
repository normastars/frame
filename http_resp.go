package frame

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response http response data
type Response struct {
	Code    string      `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Time    time.Time   `json:"time,omitempty"`
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
func (eng *Engine) Success(ctx *gin.Context, data interface{}) {
	resp := &Response{
		Code:    successCode,
		Message: successMsg,
		Data:    data,
		Time:    time.Now(),
	}
	fmt.Println(resp)

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
func (eng *Engine) Error(ctx *gin.Context, errMsg ErrorMsg) {
	resp := &Response{
		Code:    errMsg.GetCode(),
		Message: errMsg.GetReply(),
		Data:    nil,
		Time:    time.Now(),
	}
	// TODO: 打印真正的错误原因
	ctx.JSON(http.StatusOK, resp)
}

// ListSuccess 如果 pageData 是nil 或者 pageData.Results 是空,自动设置为空数组[]
// http response 结果集是数组或切片时使用
// default json
func (eng *Engine) ListSuccess(ctx *gin.Context, pageData *PageResults) {
	emptyPage(pageData)
	resp := &Response{
		Code:    successCode,
		Message: successMsg,
		Data:    pageData,
		Time:    time.Now(),
	}
	ctx.JSON(http.StatusOK, resp)
}

// ListError 自动将结果集设置为空数组
// http response 结果集是数组或切片时使用
// default json
func (eng *Engine) ListError(ctx *gin.Context, errMsg ErrorMsg) {
	resp := &Response{
		Code:    errMsg.GetCode(),
		Message: errMsg.GetReply(),
		Data:    defaultEmptyPage,
		Time:    time.Now(),
	}
	// TODO: 打印真实的错误信息
	ctx.JSON(http.StatusOK, resp)
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
