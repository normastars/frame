package frame

import (
	"net/http"
	"strconv"

	"github.com/imroc/req/v3"
	"github.com/sirupsen/logrus"
)

// ReqMetricMiddleware collects Prometheus metrics for outgoing HTTP requests.
var ReqMetricMiddleware req.ResponseMiddleware = func(c *req.Client, resp *req.Response) error {
	req := resp.Request
	code := ""
	if resp.Response != nil {
		code = strconv.Itoa(resp.Response.StatusCode)
	}
	sendHTTPRequests.WithLabelValues(
		req.Method, req.URL.Host, req.URL.Path, code,
	).Inc()
	duration := resp.TotalTime().Seconds()
	sendHTTPRequestsDuration.WithLabelValues(
		req.Method, req.URL.Host, req.URL.Path, code,
	).Observe(duration)
	return nil
}

// ReqLogMiddleware logs outgoing HTTP request/response details with trace context.
// The log entry is written on the standard logrus logger (configured via SetDefaultLog).
var ReqLogMiddleware req.ResponseMiddleware = func(c *req.Client, resp *req.Response) error {
	traceID := c.Headers.Get(TraceIDKey)
	logBody := newTraceLogFromHTTPClient(c, resp)
	logrus.WithFields(logrus.Fields{
		TraceIDKey:  traceID,
		TraceLogKey: logBody,
	}).Info("")
	return nil
}

func newTraceLogFromHTTPClient(c *req.Client, resp *req.Response) *logBody {
	cr := resp.Request
	code := 0
	if resp.Response != nil {
		code = resp.Response.StatusCode
	}

	return &logBody{
		TraceType:  TraceLogHTTPClient,
		TraceID:    c.Headers.Get(TraceIDKey),
		StatusCode: code,
		Duration:   resp.TotalTime().Milliseconds(),
		Host:       cr.URL.Host,
		Path:       cr.URL.Path,
		Extra: reqLogExtra{
			Req: reqLogBody{
				Header:      safeHTTPHeader(cr.Headers),
				PathParams:  safeStringMap(cr.PathParams),
				QueryParams: safeQueryParams(cr.QueryParams),
				Body:        requestBody(cr),
			},
			Resp: respLogBody{
				Body: safeResponseBody(resp),
			},
		},
	}
}

// safeHTTPHeader returns nil if the map is empty, so JSON omits the field.
func safeHTTPHeader(h http.Header) http.Header {
	if len(h) > 0 {
		return h
	}
	return nil
}

// safeStringMap returns nil if the map is empty, so JSON omits the field.
func safeStringMap(m map[string]string) map[string]string {
	if len(m) > 0 {
		return m
	}
	return nil
}

// safeQueryParams returns nil if the map is empty, so JSON omits the field.
func safeQueryParams(m map[string][]string) map[string][]string {
	if len(m) > 0 {
		return m
	}
	return nil
}

// requestBody returns the request body for non-GET requests, empty string otherwise.
func requestBody(cr *req.Request) string {
	if cr.Method != http.MethodGet && len(cr.Body) > 0 {
		return string(cr.Body)
	}
	return ""
}

// safeResponseBody returns the response body as a string (truncated to maxBodyLogBytes), empty on error.
func safeResponseBody(resp *req.Response) string {
	sbody, err := resp.ToString()
	if err != nil {
		return ""
	}
	return truncateBody(sbody)
}
