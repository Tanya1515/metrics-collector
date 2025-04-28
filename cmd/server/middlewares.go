package main

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	data "github.com/Tanya1515/metrics-collector.git/cmd/data"
)

// MiddlewareZipper - middleware for unpacking data from zip archive.
func (App *Application) MiddlewareZipper(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData := &ResponseData{
			status: 0,
			size:   0,
		}

		zlw := LoggingZipperResponseWriter{
			w,
			w,
			responseData,
		}
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error during comressing")
			}

			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			zlw = LoggingZipperResponseWriter{
				w,
				gz,
				responseData,
			}
		}

		next(&zlw, r)

	}
}

// MiddlewareLogger - function for logging information about processing request. 
func (App *Application) MiddlewareLogger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		method := r.Method

		responseData := &ResponseData{
			status: 0,
			size:   0,
		}

		zlw := LoggingZipperResponseWriter{
			w,
			w,
			responseData,
		}

		start := time.Now()

		h.ServeHTTP(&zlw, r)
		duration := time.Since(start)

		App.Logger.Infoln(
			"URI", uri,
			"Method", method,
			"Duration", duration,
			"ResponseStatus", responseData.status,
			"ResponseSize", responseData.size,
		)

	}
}

// MiddlewareHash - function for checking data integrity of request body. 
func (App *Application) MiddlewareHash(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseData := &ResponseData{
			status: 0,
			size:   0,
		}

		zlw := LoggingZipperResponseWriter{
			w,
			w,
			responseData,
		}

		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			App.Logger.Errorln("Error during reading request body")
			return
		}

		// Replace the body with a new reader after reading from the original
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		sign := r.Header.Get("HashSHA256")
		if sign != "" {
			signDecode, err := hex.DecodeString(sign)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				App.Logger.Errorln("Error during HashSHA256 decoding")
			}

			h := hmac.New(sha256.New, []byte(App.SecretKey))
			h.Write(body)
			signCheck := h.Sum(nil)

			if !(hmac.Equal(signDecode, signCheck)) {
				http.Error(w, "Error while checking HashSHA256 of the request", http.StatusBadRequest)
				App.Logger.Errorln("HashSHA256 is incorrect")
				return
			}
		}

		next(&zlw, r)
	}
}

// MiddlewareChain - function for processing chain of middlewares.
func (App *Application) MiddlewareChain(h http.HandlerFunc, m ...data.Middleware) http.HandlerFunc {
	for _, wrap := range m {
		h = wrap(h)
	}

	return h
}
