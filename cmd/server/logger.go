package main

import (
	"net/http"
)

type (
	ResponseData struct {
		status int
		size   int
	}

	LoggingResponseWriter struct {
		http.ResponseWriter
		responseData *ResponseData
	}
)

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
