package routerkit

import (
	"ecst-payment/pkg/tracer"
	"fmt"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TraceAndServe will apply tracing to the given http.Handler using the passed tracer under the given service and resource.
func TraceAndServe(h http.Handler, w http.ResponseWriter, r *http.Request, service, resource string, spanopts ...tracer.Option) {
	opts := append([]tracer.Option{
		{TagKey: "span.type", Value: "web"},
		{TagKey: "service.name", Value: service},
		{TagKey: "resource.name", Value: resource},
		{TagKey: "http.method", Value: r.Method},
		{TagKey: "http.url", Value: r.URL.Path},
	}, spanopts...)

	ctx, sp := tracer.SpanStartWithSpanOption(r.Context(), "http.request", opts...)

	w = wrapResponseWriter(w, sp)

	h.ServeHTTP(w, r.WithContext(ctx))
}

// wrapResponseWriter wraps an underlying http.ResponseWriter so that it can
// trace the http response codes. It also checks for various http interfaces
// (Flusher, Pusher, CloseNotifier, Hijacker) and if the underlying
// http.ResponseWriter implements them it generates an unnamed struct with the
// appropriate fields.
//
// This code is generated because we have to account for all the permutations
// of the interfaces.
func wrapResponseWriter(w http.ResponseWriter, span trace.Span) http.ResponseWriter {
	hFlusher, okFlusher := w.(http.Flusher)
	hPusher, okPusher := w.(http.Pusher)
	hCloseNotifier, okCloseNotifier := w.(http.CloseNotifier)
	hHijacker, okHijacker := w.(http.Hijacker)

	w = newResponseWriter(w, span)
	switch {
	case okFlusher && okPusher && okCloseNotifier && okHijacker:
		w = struct {
			http.ResponseWriter
			http.Flusher
			http.Pusher
			http.CloseNotifier
			http.Hijacker
		}{w, hFlusher, hPusher, hCloseNotifier, hHijacker}
	case okFlusher && okPusher && okCloseNotifier:
		w = struct {
			http.ResponseWriter
			http.Flusher
			http.Pusher
			http.CloseNotifier
		}{w, hFlusher, hPusher, hCloseNotifier}
	case okFlusher && okPusher && okHijacker:
		w = struct {
			http.ResponseWriter
			http.Flusher
			http.Pusher
			http.Hijacker
		}{w, hFlusher, hPusher, hHijacker}
	case okFlusher && okCloseNotifier && okHijacker:
		w = struct {
			http.ResponseWriter
			http.Flusher
			http.CloseNotifier
			http.Hijacker
		}{w, hFlusher, hCloseNotifier, hHijacker}
	case okPusher && okCloseNotifier && okHijacker:
		w = struct {
			http.ResponseWriter
			http.Pusher
			http.CloseNotifier
			http.Hijacker
		}{w, hPusher, hCloseNotifier, hHijacker}
	case okFlusher && okPusher:
		w = struct {
			http.ResponseWriter
			http.Flusher
			http.Pusher
		}{w, hFlusher, hPusher}
	case okFlusher && okCloseNotifier:
		w = struct {
			http.ResponseWriter
			http.Flusher
			http.CloseNotifier
		}{w, hFlusher, hCloseNotifier}
	case okFlusher && okHijacker:
		w = struct {
			http.ResponseWriter
			http.Flusher
			http.Hijacker
		}{w, hFlusher, hHijacker}
	case okPusher && okCloseNotifier:
		w = struct {
			http.ResponseWriter
			http.Pusher
			http.CloseNotifier
		}{w, hPusher, hCloseNotifier}
	case okPusher && okHijacker:
		w = struct {
			http.ResponseWriter
			http.Pusher
			http.Hijacker
		}{w, hPusher, hHijacker}
	case okCloseNotifier && okHijacker:
		w = struct {
			http.ResponseWriter
			http.CloseNotifier
			http.Hijacker
		}{w, hCloseNotifier, hHijacker}
	case okFlusher:
		w = struct {
			http.ResponseWriter
			http.Flusher
		}{w, hFlusher}
	case okPusher:
		w = struct {
			http.ResponseWriter
			http.Pusher
		}{w, hPusher}
	case okCloseNotifier:
		w = struct {
			http.ResponseWriter
			http.CloseNotifier
		}{w, hCloseNotifier}
	case okHijacker:
		w = struct {
			http.ResponseWriter
			http.Hijacker
		}{w, hHijacker}
	}

	return w
}

// responseWriter is a small wrapper around an http response writer that will
// intercept and store the status of a request.
type responseWriter struct {
	http.ResponseWriter
	span   trace.Span
	status int
}

func newResponseWriter(w http.ResponseWriter, span trace.Span) *responseWriter {
	return &responseWriter{w, span, 0}
}

// Write writes the data to the connection as part of an HTTP reply.
// We explicitely call WriteHeader with the 200 status code
// in order to get it reported into the span.
func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// WriteHeader sends an HTTP response header with status code.
// It also sets the status code to the span.
func (w *responseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
	w.span.SetAttributes(attribute.KeyValue{
		Key:   "http.status_code",
		Value: attribute.StringValue(strconv.Itoa(status)),
	})
	if status >= 500 && status < 600 {
		w.span.RecordError(fmt.Errorf("%d: %s", status, http.StatusText(status)))
	}
}
