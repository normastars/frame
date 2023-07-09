package frame

import (
	"net/http"
	"strconv"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

// var client = req.C().
// 	OnAfterResponse(ReqMetricMiddleware)

// ReqMetricMiddleware http req client
var ReqMetricMiddleware req.ResponseMiddleware = func(c *req.Client, resp *req.Response) error {
	// TODO: bus code metrics
	req := resp.Request
	code := ""
	if resp.Response != nil {
		code = strconv.Itoa(resp.Response.StatusCode)
	}
	sendHTTPRequests.WithLabelValues(
		req.Method, req.URL.Host, req.URL.Path, code,
	).Inc()
	duration := resp.TotalTime().Milliseconds()
	sendHTTPRequestsDuration.WithLabelValues(
		req.Method, req.URL.Host, req.URL.Path, code,
	).Observe(float64(duration))
	return nil
}

// ReqLogMiddleware http req client
var ReqLogMiddleware req.ResponseMiddleware = func(c *req.Client, resp *req.Response) error {
	logBody := newTraceLogFromHTTPClient(c, resp)
	l := client2logEntry(c)
	l.WithField(TraceLogKey, logBody).Info("")
	return nil
}

func newTraceLogFromHTTPClient(c *req.Client, resp *req.Response) *logBody {
	cr := resp.Request
	traceID := c.Headers.Get(TraceIDKey)
	code := 0
	if resp.Response != nil {
		code = resp.Response.StatusCode
	}
	var h http.Header
	if len(cr.Headers) > 0 {
		h = cr.Headers
	}
	var pp map[string]string
	if len(cr.PathParams) > 0 {
		pp = cr.PathParams
	}
	var qp map[string][]string
	if len(cr.QueryParams) > 0 {
		qp = cr.QueryParams
	}
	var body string
	if cr.Method != http.MethodGet && len(cr.Body) <= 0 {
		body = string(cr.Body)
	}
	sbody, _ := resp.ToString()
	return &logBody{
		TraceType:  TraceLogHTTPClient,
		TraceID:    traceID,
		StatusCode: code,
		Duration:   resp.TotalTime().Milliseconds(),
		Host:       cr.URL.Host,
		Path:       cr.URL.Path,
		Extra: reqLogExtra{
			Req: reqLogBody{
				Header:      h,
				PathParams:  pp,
				QueryParams: qp,
				Body:        body,
			},
			Resp: respLogBody{
				Body: sbody,
			},
		},
	}
}

func client2logEntry(c *req.Client) *logrus.Entry {
	traceID := c.Headers.Get(TraceIDKey)
	return NewLogger(getLogConf()).WithField(TraceIDKey, traceID)
}
