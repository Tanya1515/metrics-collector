package main

import (
	"io"
	"net/http"
)

type (
	ResponseData struct {
		status int
		size   int
	}

	LoggingZipperResponseWriter struct {
		http.ResponseWriter
		Writer       io.Writer
		responseData *ResponseData
	}
)

func (rw *LoggingZipperResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.Writer.Write(b)
	rw.responseData.size += size
	return size, err
}

func (rw *LoggingZipperResponseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.responseData.status = statusCode
}
