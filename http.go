package opentelemetry

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"opentelemetry-util/stack"
	"strconv"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TraceHTTPHandler trace http request
func TraceHTTPHandler(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(
		&traceHandler{handler},
		"operation",
		// otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		otelhttp.WithPropagators(propagation.TraceContext{}),
		otelhttp.WithFilter(func(r *http.Request) bool {
			return r.URL.Path != "/ping"
		}),
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return "Recv." + r.URL.String()
		}),
	)
}

// HTTPTransport is an http.RoundTripper that instruments all outgoing requests with OpenCensus stats and tracing.
func HTTPTransport(transport *http.Transport) http.RoundTripper {
	return otelhttp.NewTransport(
		&traceRoundTripper{transport},
		otelhttp.WithPropagators(propagation.TraceContext{}),
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			f := stack.GetFrame(10)
			path := f.File + ":" + strconv.Itoa(f.Line)
			return "Sent." + r.URL.String() + " " + path
		}),
	)
}

type traceRoundTripper struct {
	*http.Transport
}

func (rt *traceRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	span := trace.SpanFromContext(r.Context())
	if r.ContentLength > 0 {
		b, _ := ioutil.ReadAll(r.Body)
		span.SetAttributes(label.String("http.req_body", string(b)))
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}
	resp, err := rt.Transport.RoundTrip(r)
	if err != nil {
		return resp, err
	}
	if resp.ContentLength > 0 {
		b, _ := ioutil.ReadAll(resp.Body)
		span.SetAttributes(label.String("http.resp_body", string(b)))
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}
	return resp, err
}

type traceHandler struct {
	http.Handler
}

func (th *traceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	if r.ContentLength > 0 {
		b, _ := ioutil.ReadAll(r.Body)
		span.SetAttributes(label.String("http.req_body", string(b)))
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}
	tw := NewtraceResponseWriter(w)
	th.Handler.ServeHTTP(tw, r)
	respStr := tw.Body.String()
	if respStr != "" {
		span.SetAttributes(label.String("http.resp_body", respStr))
	}
}

type traceResponseWriter struct {
	http.ResponseWriter
	status int
	Body   *bytes.Buffer
}

func NewtraceResponseWriter(w http.ResponseWriter) *traceResponseWriter {
	return &traceResponseWriter{
		ResponseWriter: w,
		Body:           new(bytes.Buffer),
	}
}

func (w *traceResponseWriter) Status() int {
	return w.status
}

func (w *traceResponseWriter) Write(p []byte) (n int, err error) {
	w.Body.Write(p)
	return w.ResponseWriter.Write(p)
}

func (w *traceResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
